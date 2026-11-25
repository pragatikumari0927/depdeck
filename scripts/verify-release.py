"""Inspect a release snapshot and smoke-test a native depdeck executable."""

import argparse
import hashlib
import json
import os
import platform
import re
import subprocess
import sys
import tarfile
import tempfile
import zipfile
from pathlib import Path

VITE_LICENSE_PATH = Path(__file__).resolve().parents[1] / "pkg" / "render" / "artwork" / "LICENSE-vite.txt"
DEPDECK_MODULE_PATH = "github.com/pragatikumari0927/depdeck"


def require(condition, message):
    if not condition:
        raise SystemExit(message)


def check_go_metadata(metadata, expected_go_version, expected_module_path=None, expected_module_version=None):
    actual = metadata.get("GoVersion") if isinstance(metadata, dict) else None
    require(isinstance(actual, str) and re.fullmatch(r"go\d+\.\d+\.\d+", actual),
            f"Missing or invalid embedded Go compiler version: {actual!r}")
    if expected_go_version:
        require(actual == expected_go_version,
                f"Expected compiler {expected_go_version}, found {actual}")
    if expected_module_path is not None or expected_module_version is not None:
        main = metadata.get("Main") if isinstance(metadata, dict) else None
        observed_path = main.get("Path") if isinstance(main, dict) else None
        observed_version = main.get("Version") if isinstance(main, dict) else None
        require(observed_path == expected_module_path and observed_version == expected_module_version,
                "Expected module "
                f"{expected_module_path}@{expected_module_version}, found "
                f"{observed_path}@{observed_version}")
    return actual


def verify_go_metadata(binary, expected_go_version, expected_module_path=None, expected_module_version=None):
    result = subprocess.run(["go", "version", "-m", "-json", str(binary)],
                            capture_output=True, text=True, timeout=30)
    require(result.returncode == 0, f"Could not inspect Go build metadata for {binary}: {result.stderr}")
    try:
        metadata = json.loads(result.stdout)
    except json.JSONDecodeError as error:
        raise SystemExit(f"Invalid Go build metadata for {binary}: {error}") from error
    version = check_go_metadata(metadata, expected_go_version, expected_module_path, expected_module_version)
    if expected_module_path is not None:
        main = metadata.get("Main")
        observed_path = main.get("Path") if isinstance(main, dict) else None
        observed_version = main.get("Version") if isinstance(main, dict) else None
        print(f"Requested module {expected_module_path}@{expected_module_version}")
        print(f"Embedded module {observed_path}@{observed_version}  {binary}", flush=True)
    print(f"Embedded compiler {version}  {binary}", flush=True)
    return version


def check_application_identity(observed, expected_version, expected_module_path=None, expected_module_version=None):
    canonical = f"depdeck {expected_version}"
    legacy_v1_identity = (
        expected_version == "1.0.0"
        and expected_module_path == DEPDECK_MODULE_PATH
        and expected_module_version == "v1.0.0"
        and observed == "depdeck v1.0.0"
    )
    if legacy_v1_identity:
        return "v1.0.0"
    require(observed == canonical,
            f"Expected application identity {canonical!r}, observed {observed!r}")
    return expected_version


def check_mcp_server_version(observed, expected_version):
    require(observed == expected_version,
            f"Expected MCP server version {expected_version!r}, observed {observed!r}")


def check_binary_scan_result(result, binary):
    require(result.returncode == 0,
            f"govulncheck binary scan failed for {binary}: {result.stdout}{result.stderr}")
    print(f"Binary vulnerability scan passed  {binary}")


def check_artwork_notice(notice, require_vite=True):
    if not require_vite:
        return
    required = (
        "Package artwork: Vite",
        "vitejs/vite/docs/public/logo.svg",
        "Copyright (c) 2019-present, VoidZero Inc. and Vite contributors",
        "Permission is hereby granted, free of charge",
        "imply affiliation with or endorsement by Vite, VoidZero, or its contributors",
    )
    for item in required:
        require(item in notice, f"Third-party notices are missing artwork attribution: {item}")
    license_text = VITE_LICENSE_PATH.read_text(encoding="utf-8").strip()
    require(license_text in notice, "Third-party notices do not contain the complete Vite MIT license")


def scan_binary(scanner, binary):
    result = subprocess.run([str(scanner), "-mode=binary", str(binary)],
                            capture_output=True, text=True, timeout=120)
    check_binary_scan_result(result, binary)


def expected_mcp_tools(expect_render_deck):
    tools = {"list_deck", "get_card", "check_deck"}
    if expect_render_deck:
        tools.add("render_deck")
    return tools


def check_mcp_tools(actual_tools, expect_render_deck):
    expected = expected_mcp_tools(expect_render_deck)
    require(actual_tools == expected, f"Unexpected MCP tool inventory: {sorted(actual_tools)}")


def check_envelope(envelope, tool, identity):
    require(isinstance(envelope, dict), f"Invalid {tool} JSON envelope")
    schema_version = envelope.get("schema_version")
    require(isinstance(schema_version, str) and schema_version.split(".", 1)[0] == "1",
            f"Unexpected {tool} JSON schema version: {schema_version!r}")
    require(envelope.get("generated_by") == identity and envelope.get("tool") == tool,
            f"Unexpected {tool} JSON metadata")
    require(isinstance(envelope.get("data"), dict), f"Missing {tool} JSON data")
    return envelope["data"]


def check_empty_check_data(data, require_empty_array):
    require(data.get("reason") == "empty_roster", "Unexpected empty Check reason")
    if "dependencies" in data:
        require(data["dependencies"] == [], "Empty Check dependencies must be an empty array")
    else:
        require(not require_empty_array, "Empty Check must include data.dependencies: []")


def check_render_result(envelope, identity, output, actual_bytes, html):
    data = check_envelope(envelope, "render_deck", identity)
    require(data.get("output") == output and data.get("cards") == 0 and
            data.get("bytes") == actual_bytes and actual_bytes > 0,
            "Unexpected MCP Deck render result")
    require("Rarity Atlas" in html, "MCP Deck render did not use the canonical renderer")


def smoke(binary, version, expect_render_deck=False, expected_go_version=None,
          govulncheck=None, binary_checks_done=False, expected_module_path=None,
          expected_module_version=None):
    if not binary_checks_done:
        verify_go_metadata(binary, expected_go_version, expected_module_path, expected_module_version)
        if govulncheck:
            scan_binary(govulncheck, binary)
    with tempfile.TemporaryDirectory(prefix="depdeck-smoke-") as directory:
        root = Path(directory)
        (root / "package.json").write_text("{}", encoding="utf-8")
        env = dict(os.environ, HOME=directory, USERPROFILE=directory,
                   LOCALAPPDATA=directory, XDG_CACHE_HOME=directory,
                   HTTPS_PROXY="http://127.0.0.1:1", HTTP_PROXY="http://127.0.0.1:1", NO_PROXY="")

        def run(*args, stdin=None, code=0):
            result = subprocess.run([binary, *args], input=stdin, capture_output=True,
                                    text=True, cwd=root, env=env, timeout=30)
            require(result.returncode == code, f"{args} returned {result.returncode}: {result.stderr}")
            return result

        app_version = check_application_identity(
            run("--version").stdout.strip(), version, expected_module_path, expected_module_version
        )
        identity = f"depdeck {app_version}"
        require(run("scan", "--flavor=ai", code=2).stderr, "Missing CLI error diagnostic")
        cli = json.loads(run("check", "--json").stdout)
        cli_data = check_envelope(cli, "check", identity)
        check_empty_check_data(cli_data, expect_render_deck)

        process = subprocess.Popen([binary, "mcp"], stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE, text=True, cwd=root, env=env)
        with process:
            try:
                def send(message):
                    process.stdin.write(json.dumps({"jsonrpc": "2.0", **message}) + "\n")
                    process.stdin.flush()

                def receive():
                    import threading
                    expired = threading.Timer(30, process.kill)
                    expired.start()
                    try:
                        return json.loads(process.stdout.readline())
                    finally:
                        expired.cancel()

                send({"id": 1, "method": "initialize", "params": {
                    "protocolVersion": "2025-11-25", "capabilities": {},
                    "clientInfo": {"name": "release-smoke", "version": "1"},
                }})
                check_mcp_server_version(receive()["result"]["serverInfo"]["version"], app_version)
                send({"method": "notifications/initialized"})
                send({"id": 2, "method": "tools/list"})
                tools = receive()["result"]["tools"]
                check_mcp_tools({tool["name"] for tool in tools}, expect_render_deck)
                send({"id": 3, "method": "tools/call", "params": {
                    "name": "check_deck", "arguments": {"path": "."},
                }})
                assessment = receive()["result"]["structuredContent"]
                check_data = check_envelope(assessment, "check_deck", identity)
                check_empty_check_data(check_data, expect_render_deck)
                if expect_render_deck:
                    send({"id": 4, "method": "tools/call", "params": {
                        "name": "render_deck", "arguments": {"path": ".", "out": "release-smoke.html"},
                    }})
                    rendered = receive()["result"]["structuredContent"]
                    deck_bytes = (root / "release-smoke.html").read_bytes()
                    deck = deck_bytes.decode("utf-8")
                    check_render_result(rendered, identity, "release-smoke.html", len(deck_bytes), deck)
                process.stdin.close()
                require(process.wait(timeout=30) == 0, "MCP failed to shut down on EOF")
                require(not process.stderr.read(), "Unexpected MCP diagnostic")
            finally:
                if process.poll() is None:
                    process.kill()
    print(f"CLI and MCP smoke passed for {identity}")


def verify(dist, expected_version, expect_render_deck=False, expected_go_version=None, govulncheck=None):
    expected = {(os_name, arch) for os_name in ("windows", "linux", "darwin")
                for arch in ("amd64", "arm64")}
    seen = set()
    native = ({"win32": "windows"}.get(sys.platform, sys.platform),
              {"x86_64": "amd64", "AMD64": "amd64", "aarch64": "arm64", "ARM64": "arm64"}
              .get(platform.machine(), platform.machine()))
    native_checked = False
    for line in (dist / "checksums.txt").read_text(encoding="utf-8").splitlines():
        digest, name = line.split()
        match = re.fullmatch(r"depdeck_(.+)_(windows|linux|darwin)_(amd64|arm64)\.(zip|tar\.gz)", name)
        require(match, f"Unexpected archive name: {name}")
        version, os_name, arch, extension = match.groups()
        require(version == expected_version,
                f"Archive {name} is version {version}, expected approved version {expected_version}")
        target = (os_name, arch)
        require(target not in seen, f"Duplicate target: {target}")
        seen.add(target)
        archive = dist / name
        require(hashlib.sha256(archive.read_bytes()).hexdigest() == digest, f"SHA-256 mismatch: {name}")
        require((extension == "zip") == (os_name == "windows"), f"Wrong archive format: {name}")
        executable = "depdeck.exe" if os_name == "windows" else "depdeck"
        members = {executable, "LICENSE", "THIRD_PARTY_NOTICES.txt", "README.md", "CHANGELOG.md"}
        with (zipfile.ZipFile(archive) if extension == "zip" else tarfile.open(archive)) as package:
            names = package.namelist() if extension == "zip" else package.getnames()
            require(len(names) == len(members) and set(names) == members, f"Wrong contents: {name}: {names}")
            notice_bytes = (package.read("THIRD_PARTY_NOTICES.txt") if extension == "zip"
                            else package.extractfile("THIRD_PARTY_NOTICES.txt").read())
            check_artwork_notice(notice_bytes.decode("utf-8"), expect_render_deck)
            with tempfile.TemporaryDirectory(prefix="depdeck-archive-") as directory:
                binary = Path(directory) / executable
                if extension == "zip":
                    binary.write_bytes(package.read(executable))
                else:
                    with package.extractfile(executable) as stream:
                        binary.write_bytes(stream.read())
                binary.chmod(0o755)
                verify_go_metadata(binary, expected_go_version)
                if govulncheck:
                    scan_binary(govulncheck, binary)
                if target == native:
                    smoke(binary, version, expect_render_deck, expected_go_version,
                          govulncheck, binary_checks_done=True)
                    native_checked = True
        print(f"{digest}  {name}")
    require(seen == expected, f"Missing archive targets: {expected - seen}")
    require(native_checked, f"No archive was smoke-tested for this host: {native}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("dist", nargs="?", type=Path)
    parser.add_argument("--binary", type=Path)
    parser.add_argument("--version", help="approved release version for archives or expected binary version")
    parser.add_argument("--expect-render-deck", action="store_true",
                        help="require and exercise the Pixel Archive MCP render_deck tool")
    parser.add_argument("--expected-go-version",
                        help="require this embedded compiler version, for example go1.27.4")
    parser.add_argument("--module-path", help="require this embedded Go module path")
    parser.add_argument("--module-version", help="require this embedded Go module version")
    parser.add_argument("--govulncheck", type=Path,
                        help="run govulncheck binary mode on every inspected executable")
    args = parser.parse_args()
    if (args.module_path is None) != (args.module_version is None):
        parser.error("--module-path and --module-version must be provided together")
    if args.binary and args.version and not args.dist:
        smoke(args.binary.resolve(), args.version, args.expect_render_deck,
              args.expected_go_version, args.govulncheck, expected_module_path=args.module_path,
              expected_module_version=args.module_version)
    elif args.dist and args.version and not args.binary:
        verify(args.dist, args.version, args.expect_render_deck,
               args.expected_go_version, args.govulncheck)
    else:
        parser.error("provide a snapshot directory and --version, or --binary and --version")
