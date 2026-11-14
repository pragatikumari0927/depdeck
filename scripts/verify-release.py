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


def require(condition, message):
    if not condition:
        raise SystemExit(message)


def smoke(binary, version):
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

        identity = f"depdeck {version}"
        require(run("--version").stdout.strip() == identity, "Unexpected CLI version")
        require(run("scan", "--flavor=ai", code=2).stderr, "Missing CLI error diagnostic")
        cli = json.loads(run("check", "--json").stdout)
        require(cli["generated_by"] == identity and cli["data"]["reason"] == "empty_roster",
                "Unexpected CLI assessment")

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
                require(receive()["result"]["serverInfo"]["version"] == version, "Unexpected MCP version")
                send({"method": "notifications/initialized"})
                send({"id": 2, "method": "tools/list"})
                tools = receive()["result"]["tools"]
                require({tool["name"] for tool in tools} == {"list_deck", "get_card", "check_deck"},
                        "Unexpected MCP tool inventory")
                send({"id": 3, "method": "tools/call", "params": {
                    "name": "check_deck", "arguments": {"path": "."},
                }})
                assessment = receive()["result"]["structuredContent"]
                require(assessment["generated_by"] == identity and assessment["data"]["reason"] == "empty_roster",
                        "Unexpected MCP assessment")
                process.stdin.close()
                require(process.wait(timeout=30) == 0, "MCP failed to shut down on EOF")
                require(not process.stderr.read(), "Unexpected MCP diagnostic")
            finally:
                if process.poll() is None:
                    process.kill()
    print(f"CLI and MCP smoke passed for {identity}")


def verify(dist):
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
            if target == native:
                with tempfile.TemporaryDirectory(prefix="depdeck-archive-") as directory:
                    binary = Path(directory) / executable
                    if extension == "zip":
                        binary.write_bytes(package.read(executable))
                    else:
                        with package.extractfile(executable) as stream:
                            binary.write_bytes(stream.read())
                    binary.chmod(0o755)
                    smoke(binary, version)
                    native_checked = True
        print(f"{digest}  {name}")
    require(seen == expected, f"Missing archive targets: {expected - seen}")
    require(native_checked, f"No archive was smoke-tested for this host: {native}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("dist", nargs="?", type=Path)
    parser.add_argument("--binary", type=Path)
    parser.add_argument("--version")
    args = parser.parse_args()
    if args.binary and args.version and not args.dist:
        smoke(args.binary.resolve(), args.version)
    elif args.dist and not args.binary and not args.version:
        verify(args.dist)
    else:
        parser.error("provide a snapshot directory or --binary and --version")