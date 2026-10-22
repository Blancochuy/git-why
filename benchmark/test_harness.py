import hashlib
import json
import os
from pathlib import Path
import tempfile
import unittest

from harness import Broker, CASES, dry_run, evaluate, fixture, git, grade


class HarnessTests(unittest.TestCase):
    def test_histories_are_deterministic_and_identical_between_conditions(self):
        for case_id in CASES:
            with self.subTest(case=case_id), tempfile.TemporaryDirectory() as temp:
                parent = Path(temp)
                shas, snapshots = [], []
                for condition in "ABC":
                    location = parent / condition
                    location.mkdir()
                    repo, history = fixture(location, case_id, condition)
                    shas.append(history)
                    snapshots.append(hashlib.sha256((repo / "subject.py").read_bytes()).hexdigest())
                    self.assertEqual((repo / ".git").exists(), condition != "A")
                    if condition != "A":
                        self.assertEqual(git(repo, "remote"), "")
                        self.assertEqual(len(git(repo, "rev-list", "--all").splitlines()), len(history))
                        self.assertEqual(git(repo, "branch", "--format=%(refname)"), "refs/heads/main")
                    self.assertEqual({p.name for p in repo.iterdir()} - {".git"}, {"subject.py"})
                self.assertEqual(shas[0], shas[1])
                self.assertEqual(shas[1], shas[2])
                self.assertEqual(len(set(snapshots)), 1)

    def test_reference_passes_mutant_and_original_fail(self):
        for case_id, case in CASES.items():
            if "reference" not in case:
                continue
            with self.subTest(case=case_id):
                self.assertTrue(grade(case_id, case["reference"])["task_success"])
                self.assertFalse(grade(case_id, case["mutant"])["task_success"])
                # Regression checks alone must not accept an unimplemented feature.
                self.assertFalse(grade(case_id, case["history"][-1][1])["task_success"])

    def test_unknown_requires_human_review(self):
        result = grade("unknown", CASES["unknown"]["history"][-1][1])
        self.assertIsNone(result["task_success"])
        self.assertEqual(result["status"], "human_review_required")

    def test_no_code_execution(self):
        malicious = [
            "import os\ndef retry_delay(seconds):\n    return 0\n",
            "def retry_delay(seconds):\n    return __import__('os').getcwd()\n",
            "@print('side effect')\ndef retry_delay(seconds):\n    return 0\n",
            "def retry_delay(seconds):\n    while True:\n        pass\n",
        ]
        for source in malicious:
            with self.subTest(source=source):
                self.assertEqual(grade("fallback", source)["status"], "unsupported_or_invalid_patch")

    def test_broker_denies_paths_shell_network_and_history_in_a(self):
        with tempfile.TemporaryDirectory() as temp:
            repo, _ = fixture(Path(temp), "fallback", "A")
            broker = Broker(repo, "A")
            for tool, args in [
                ("read", {"file": "../reference.py"}),
                ("read", {"file": ".git/config"}),
                ("read", {"file": str(Path(temp) / "private.json")}),
                ("write", {"file": "../escape.py", "source": ""}),
                ("shell", {"command": "git log"}),
                ("history", {}), ("blame", {}),
                ("fetch", {"url": "https://example.invalid"}),
                ("context", {"file": "subject.py", "line": 1}),
            ]:
                self.assertIsNotNone(broker.call(tool, **args)["error"])
            self.assertIsNone(broker.call("read", file="subject.py")["error"])
            self.assertFalse((Path(temp) / "escape.py").exists())

    def test_limits_and_errors_are_recorded(self):
        with tempfile.TemporaryDirectory() as temp:
            repo, _ = fixture(Path(temp), "fallback", "B")
            broker = Broker(repo, "B", max_calls=1)
            self.assertIsNone(broker.call("read", file="subject.py")["error"])
            self.assertIn("budget", broker.call("history")["error"])
            expired = Broker(repo, "B", deadline_seconds=0)
            self.assertIn("deadline", expired.call("history")["error"])
            self.assertEqual(len(broker.events), 2)

    def test_git_patch_parity_and_revision_restrictions(self):
        for condition in "BC":
            with self.subTest(condition=condition), tempfile.TemporaryDirectory() as temp:
                repo, shas = fixture(Path(temp), "fallback", condition)
                broker = Broker(repo, condition)
                patch = broker.call("show", sha=shas[-1])
                self.assertIsNone(patch["error"])
                self.assertIn("+    if seconds <= 0:", patch["result"])
                self.assertIsNotNone(broker.call("show", sha="--all")["error"])
                self.assertIsNotNone(broker.call("show", sha="f" * 40)["error"])

    def test_earlier_rationale_requires_more_than_latest_commit(self):
        with tempfile.TemporaryDirectory() as temp:
            repo, shas = fixture(Path(temp), "earlier", "B")
            self.assertNotIn("Downstream", git(repo, "log", "-1", "--format=%B"))
            self.assertIn("Downstream", git(repo, "show", "-s", "--format=%B", shas[1]))
            self.assertIn(shas[-1], git(repo, "blame", "--porcelain", "-L4,4", "subject.py"))

    def test_self_check_preserves_all_attempts_and_never_overwrites(self):
        with tempfile.TemporaryDirectory() as temp:
            output = Path(temp) / "results"
            result = dry_run(output)
            self.assertEqual(len(result["attempts"]), 9)
            self.assertIn("not_model_evaluation", result["kind"])
            self.assertIsNone(result["model"])
            self.assertEqual(len(json.loads((output / "self-check.json").read_text())["attempts"]), 9)
            with self.assertRaises(FileExistsError):
                dry_run(output)

    @unittest.skipUnless(os.environ.get("GITWHY_BENCH_BINARY"), "Set GITWHY_BENCH_BINARY for real MCP smoke test")
    def test_real_mcp_condition_c(self):
        with tempfile.TemporaryDirectory() as temp:
            repo, shas = fixture(Path(temp), "fallback", "C")
            broker = Broker(repo, "C", binary=os.environ["GITWHY_BENCH_BINARY"])
            response = broker.call("context", file="subject.py", line=2)
            self.assertIsNone(response["error"])
            self.assertIn(shas[-1][:12], response["result"])
            self.assertIn("disables retries", response["result"])
            self.assertIsNotNone(broker.call("context", file="../subject.py", line=2)["error"])


if __name__ == "__main__":
    unittest.main()
