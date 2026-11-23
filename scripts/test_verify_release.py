import importlib.util
from types import SimpleNamespace
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("verify-release.py")
SPEC = importlib.util.spec_from_file_location("verify_release", SCRIPT)
verify_release = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(verify_release)


class MCPReleaseSmokeTests(unittest.TestCase):
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
