import unittest

from scripts.check_quality_gate import check_results


class QualityGateTests(unittest.TestCase):
    def test_all_required_jobs_must_succeed(self):
        check_results({"test": "success", "browser": "success", "security": "success"})

    def test_failed_skipped_cancelled_or_missing_jobs_fail(self):
        for state in ("failure", "skipped", "cancelled", "missing"):
            with self.subTest(state=state), self.assertRaises(SystemExit):
                check_results({"test": "success", "browser": state, "security": "success"})

    def test_each_required_job_is_enforced(self):
        for job in ("test", "browser", "security"):
            results = {name: "success" for name in ("test", "browser", "security")}
            results[job] = "failure"
            with self.subTest(job=job), self.assertRaises(SystemExit):
                check_results(results)


if __name__ == "__main__":
    unittest.main()
