"""Fail unless every required quality job completed successfully."""

import sys


REQUIRED_JOBS = ("test", "browser", "security")


def check_results(results):
    failures = {name: results.get(name, "missing") for name in REQUIRED_JOBS
                if results.get(name) != "success"}
    if failures:
        summary = ", ".join(f"{name}={result}" for name, result in failures.items())
        raise SystemExit(f"Required quality jobs did not pass: {summary}")


if __name__ == "__main__":
    if len(sys.argv) != len(REQUIRED_JOBS) + 1:
        raise SystemExit("usage: check_quality_gate.py TEST BROWSER SECURITY")
    check_results(dict(zip(REQUIRED_JOBS, sys.argv[1:])))
    print("All required quality jobs passed.")
