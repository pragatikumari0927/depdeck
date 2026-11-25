import hashlib
import io
import importlib.util
import tarfile
import tempfile
import zipfile
from types import SimpleNamespace
import unittest
from pathlib import Path
from unittest import mock


SCRIPT = Path(__file__).with_name("verify-release.py")
SPEC = importlib.util.spec_from_file_location("verify_release", SCRIPT)
verify_release = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(verify_release)


class MCPReleaseSmokeTests(unittest.TestCase):
    def create_archive_set(self, dist, version_for_target, notice, default_version="1.1.0"):
        entries = []
        for os_name, arch in sorted(
            (os_name, arch)
            for os_name in ("windows", "linux", "darwin")
            for arch in ("amd64", "arm64")
        ):
            version = version_for_target.get((os_name, arch), default_version)
            extension = "zip" if os_name == "windows" else "tar.gz"
            name = f"depdeck_{version}_{os_name}_{arch}.{extension}"
            binary = "depdeck.exe" if os_name == "windows" else "depdeck"
            files = {
                binary: b"test executable",
                "LICENSE": b"MIT",
                "THIRD_PARTY_NOTICES.txt": notice.encode("utf-8"),
                "README.md": b"README",
                "CHANGELOG.md": b"CHANGELOG",
            }
            archive = dist / name
            if extension == "zip":
                with zipfile.ZipFile(archive, "w") as package:
                    for member, payload in files.items():
                        package.writestr(member, payload)
            else:
                with tarfile.open(archive, "w:gz") as package:
                    for member, payload in files.items():
                        info = tarfile.TarInfo(member)
                        info.size = len(payload)
                        package.addfile(info, io.BytesIO(payload))
            entries.append((hashlib.sha256(archive.read_bytes()).hexdigest(), name))
        (dist / "checksums.txt").write_text(
            "".join(f"{digest}  {name}\n" for digest, name in entries), encoding="utf-8"
        )

    def test_repository_third_party_notices_include_complete_vite_license(self):
        root_notices = verify_release.VITE_LICENSE_PATH.parents[3] / "THIRD_PARTY_NOTICES.txt"
        verify_release.check_artwork_notice(root_notices.read_text(encoding="utf-8"))

    def test_artwork_notice_requires_vite_provenance_license_and_trademark_scope(self):
        notice = """Package artwork: Vite
Source: vitejs/vite/docs/public/logo.svg
License: MIT License
Copyright (c) 2019-present, VoidZero Inc. and Vite contributors
Permission is hereby granted, free of charge
The Vite mark identifies the exact npm package. This use does not imply affiliation with or endorsement by Vite, VoidZero, or its contributors.
"""
        license_text = verify_release.VITE_LICENSE_PATH.read_text(encoding="utf-8")
        verify_release.check_artwork_notice(notice + license_text)

    def test_artwork_notice_rejects_missing_license_or_provenance(self):
        for notice in (
            "Package artwork: Vite\nSource: vitejs/vite/docs/public/logo.svg\n",
            "Package artwork: Vite\nLicense: MIT License\n",
            "Package artwork: Vite\nSource: vitejs/vite/docs/public/logo.svg\nLicense: MIT License\n",
        ):
            with self.subTest(notice=notice), self.assertRaises(SystemExit):
                verify_release.check_artwork_notice(notice)

    def test_archives_must_include_vite_notice(self):
        # Archive-level extraction uses the same notice validator as the public helper.
        with self.assertRaises(SystemExit):
            verify_release.check_artwork_notice("Depdeck third-party notices\n")

    def test_legacy_release_notice_without_package_artwork_is_accepted(self):
        # v1.0.0 predates curated package artwork and has no Vite attribution.
        with tempfile.TemporaryDirectory() as directory:
            dist = Path(directory)
            notice = "Depdeck third-party notices\nNo curated package marks are included.\n"
            self.create_archive_set(dist, {}, notice, default_version="1.0.0")
            with mock.patch.object(verify_release, "verify_go_metadata"), \
                 mock.patch.object(verify_release, "smoke"):
                verify_release.verify(dist, "1.0.0", expect_render_deck=False)

    def test_archive_set_must_match_an_explicit_approved_version(self):
        with tempfile.TemporaryDirectory() as directory:
            dist = Path(directory)
            self.create_archive_set(
                dist, {("linux", "arm64"): "1.0.0"}, "Placeholder notice"
            )
            with mock.patch.object(verify_release, "check_artwork_notice"), \
                 mock.patch.object(verify_release, "verify_go_metadata"), \
                 mock.patch.object(verify_release, "smoke"):
                with self.assertRaisesRegex(SystemExit, "expected approved version 1.1.0"):
                    verify_release.verify(dist, "1.1.0")

    def test_v1_release_requires_original_three_tools(self):
        self.assertEqual(
            verify_release.expected_mcp_tools(False),
            {"list_deck", "get_card", "check_deck"},
        )

    def test_pixel_archive_release_requires_render_tool(self):
        self.assertEqual(
            verify_release.expected_mcp_tools(True),
            {"list_deck", "get_card", "check_deck", "render_deck"},
        )

    def test_missing_or_unexpected_tools_fail(self):
        with self.assertRaises(SystemExit):
            verify_release.check_mcp_tools({"list_deck", "get_card"}, False)
        with self.assertRaises(SystemExit):
            verify_release.check_mcp_tools(
                {"list_deck", "get_card", "check_deck", "render_deck"}, False
            )

    def test_empty_check_accepts_v1_compatibility_and_requires_current_empty_array(self):
        verify_release.check_empty_check_data({"reason": "empty_roster"}, False)
        verify_release.check_empty_check_data({"reason": "empty_roster", "dependencies": []}, True)
        with self.assertRaises(SystemExit):
            verify_release.check_empty_check_data({"reason": "empty_roster"}, True)
        with self.assertRaises(SystemExit):
            verify_release.check_empty_check_data({"reason": "empty_roster", "dependencies": None}, False)

    def test_compiler_metadata_requires_go_version_and_matches_requested_patch(self):
        self.assertEqual(verify_release.check_go_metadata({"GoVersion": "go1.27.4"}, "go1.27.4"), "go1.27.4")
        with self.assertRaises(SystemExit):
            verify_release.check_go_metadata({}, None)
        with self.assertRaises(SystemExit):
            verify_release.check_go_metadata({"GoVersion": "go1.27.3"}, "go1.27.4")

    def test_go_module_identity_requires_exact_path_and_requested_tag(self):
        metadata = {
            "GoVersion": "go1.27.4",
            "Main": {"Path": "github.com/pragatikumari0927/depdeck", "Version": "v1.1.0"},
        }
        self.assertEqual(
            verify_release.check_go_metadata(
                metadata, "go1.27.4", "github.com/pragatikumari0927/depdeck", "v1.1.0"
            ),
            "go1.27.4",
        )
        for changed in (
            {**metadata, "Main": {"Path": "example.com/depdeck", "Version": "v1.1.0"}},
            {**metadata, "Main": {"Path": "github.com/pragatikumari0927/depdeck", "Version": "v1.0.0"}},
        ):
            with self.subTest(main=changed["Main"]), self.assertRaisesRegex(SystemExit, "Expected module"):
                verify_release.check_go_metadata(
                    changed, "go1.27.4", "github.com/pragatikumari0927/depdeck", "v1.1.0"
                )

    def test_v1_0_0_legacy_identity_is_narrowly_accepted(self):
        self.assertEqual(
            verify_release.check_application_identity(
                "depdeck v1.0.0", "1.0.0", "github.com/pragatikumari0927/depdeck", "v1.0.0"
            ),
            "v1.0.0",
        )
        with self.assertRaisesRegex(SystemExit, "Expected application identity 'depdeck 1.1.0'"):
            verify_release.check_application_identity(
                "depdeck v1.0.0", "1.1.0", "github.com/pragatikumari0927/depdeck", "v1.1.0"
            )
        with self.assertRaisesRegex(SystemExit, "Expected application identity 'depdeck 1.0.0'"):
            verify_release.check_application_identity(
                "depdeck v1.0.0", "1.0.0", "example.com/depdeck", "v1.0.0"
            )

    def test_canonical_version_identity_is_required_for_future_tag(self):
        self.assertEqual(
            verify_release.check_application_identity(
                "depdeck 1.1.0", "1.1.0", "github.com/pragatikumari0927/depdeck", "v1.1.0"
            ),
            "1.1.0",
        )
        for observed in ("depdeck dev", "depdeck 1.0.0", "depdeck v1.1.0"):
            with self.subTest(observed=observed), self.assertRaises(SystemExit):
                verify_release.check_application_identity(
                    observed, "1.1.0", "github.com/pragatikumari0927/depdeck", "v1.1.0"
                )

    def test_mcp_version_must_match_application_identity(self):
        verify_release.check_mcp_server_version("v1.0.0", "v1.0.0")
        verify_release.check_mcp_server_version("1.1.0", "1.1.0")
        with self.assertRaisesRegex(SystemExit, "Expected MCP server version '1.1.0'.*observed 'dev'"):
            verify_release.check_mcp_server_version("dev", "1.1.0")

    def test_binary_scan_must_pass(self):
        verify_release.check_binary_scan_result(SimpleNamespace(returncode=0, stdout="clean", stderr=""), "depdeck")
        with self.assertRaises(SystemExit):
            verify_release.check_binary_scan_result(
                SimpleNamespace(returncode=1, stdout="found vulnerability", stderr=""), "depdeck"
            )

    def test_render_deck_result_and_html_are_verified(self):
        verify_release.check_render_result(
            {
                "schema_version": "1.1",
                "generated_by": "depdeck 1.1.0",
                "tool": "render_deck",
                "data": {"output": "release-smoke.html", "cards": 0, "bytes": 64},
            },
            "depdeck 1.1.0",
            "release-smoke.html",
            64,
            "<h2>Rarity Atlas</h2>",
        )

    def test_render_deck_rejects_wrong_result_or_missing_atlas(self):
        result = {
            "schema_version": "1.1",
            "generated_by": "depdeck 1.1.0",
            "tool": "render_deck",
            "data": {"output": "release-smoke.html", "cards": 1, "bytes": 64},
        }
        with self.assertRaises(SystemExit):
            verify_release.check_render_result(
                result, "depdeck 1.1.0", "release-smoke.html", 64, "Rarity Atlas"
            )
        result["data"]["cards"] = 0
        with self.assertRaises(SystemExit):
            verify_release.check_render_result(
                result, "depdeck 1.1.0", "release-smoke.html", 64, "old card template"
            )


if __name__ == "__main__":
    unittest.main()
