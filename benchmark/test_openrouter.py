import json
import os
from pathlib import Path
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch
import urllib.error

from openrouter_runner import (API, RateLimited, RunError, execute, free_tool_model, load_env, make_plan,
                               payload_for, redact, retry_seconds, run_attempt, save, select_models, summarize, tools_for)
from harness import CASES


class FakeAPI:
    def __init__(self, responses):
        self.responses = iter(responses)
        self.payloads = []

    def request(self, endpoint, payload, timeout):
        self.payloads.append(json.loads(json.dumps(payload)))
        response = next(self.responses)
        if isinstance(response, Exception):
            raise response
        return response


def response(content="Done", calls=None, finish="stop", **extra):
    return {"model": "test/model:free", "provider": "fixture-provider", "id": "fake-generation",
            "choices": [{"finish_reason": finish, "message": {"role": "assistant",
                         "content": content, **({"tool_calls": calls} if calls else {})}}],
            "usage": {"prompt_tokens": 12, "completion_tokens": 4, "cost": 0}, **extra}


def call(name, args, identity="call-1"):
    return {"id": identity, "type": "function", "function": {
        "name": name, "arguments": json.dumps(args)}}


class OpenRouterTests(unittest.TestCase):
    def settings(self):
        return {"max_tools": 20, "seconds": 30, "binary": None,
                "max_turns": 8, "interval": 0, "max_tokens": 1000}

    def run_fake(self, fake, condition="A", settings=None):
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp)
            result = run_attempt({"model": "test/model:free", "case": "fallback", "condition": condition},
                                 settings or self.settings(), directory, "test-secret", api=fake)
            saved = (directory / "attempt.json").read_text(encoding="utf-8")
            self.assertNotIn("test-secret", saved)
            return result

    def test_free_filter_rejects_paid_router_missing_and_negative_prices(self):
        good = {"id": "test/model:free", "pricing": {"prompt": "0", "completion": "0", "request": "0"},
                "supported_parameters": ["tools", "temperature", "max_tokens"]}
        self.assertTrue(free_tool_model(good))
        for change in [{"id": "openrouter/free"}, {"id": "paid/model"},
                       {"pricing": {"prompt": "0", "completion": "1"}},
                       {"pricing": {"prompt": "0", "completion": "0", "request": "-1"}},
                       {"pricing": {}}, {"supported_parameters": []}]:
            self.assertFalse(free_tool_model(dict(good, **change)))
        with self.assertRaises(RunError):
            select_models({"data": [good]}, ["unknown/model:free"])

    def test_payload_has_free_only_cost_caps_no_fallback(self):
        payload = payload_for("test/model:free", [], "C", 1000)
        self.assertNotIn("models", payload)
        self.assertFalse(payload["provider"]["allow_fallbacks"])
        self.assertTrue(payload["provider"]["require_parameters"])
        self.assertEqual(payload["provider"]["max_price"], {"prompt": 0, "completion": 0, "request": 0})
        with self.assertRaises(RunError):
            payload_for("test/model", [], "C", 1000)

    def test_tools_match_conditions(self):
        names = lambda c: {tool["function"]["name"] for tool in tools_for(c)}
        self.assertEqual(names("A"), {"read", "write"})
        self.assertEqual(names("B") - names("A"), {"history", "blame", "show"})
        self.assertEqual(names("C") - names("B"), {"context"})

    def test_plan_balanced_seeded_and_fresh_ids(self):
        plan = make_plan(["test/model:free"], "pilot", 42)
        self.assertEqual(plan, make_plan(["test/model:free"], "pilot", 42))
        self.assertEqual(len(plan), 9 * len(CASES))
        self.assertEqual(len({p["id"] for p in plan}), 9 * len(CASES))
        for case in CASES:
            for condition in "ABC":
                self.assertEqual(sum(p["case"] == case and p["condition"] == condition for p in plan), 3)

    def test_exploratory_subset_keeps_balanced_conditions(self):
        plan = make_plan(["test/model:free"], "exploratory", 43, ["reserved_signal", "unknown"], 1)
        self.assertEqual(len(plan), 6)
        self.assertEqual({p["case"] for p in plan}, {"reserved_signal", "unknown"})
        for case in ("reserved_signal", "unknown"):
            self.assertEqual({p["condition"] for p in plan if p["case"] == case}, set("ABC"))

    def test_dotenv_is_literal_and_does_not_override_environment(self):
        with tempfile.TemporaryDirectory() as temp, patch.dict(os.environ, {"OPENROUTER_API_KEY": "already-set"}, clear=True):
            path = Path(temp) / ".env"
            path.write_text('OPENROUTER_API_KEY="file-key"\nOPENROUTER_MODEL=$(touch nope)\n', encoding="utf-8-sig")
            load_env(path)
            self.assertEqual(os.environ["OPENROUTER_API_KEY"], "already-set")
            self.assertEqual(os.environ["OPENROUTER_MODEL"], "$(touch nope)")

    def test_secret_redaction(self):
        self.assertEqual(redact({"x": "secret sk-or-v1-example"}, "secret"), {"x": "[REDACTED] [REDACTED]"})

    def test_model_calls_tools_edits_and_finishes_without_private_answers(self):
        fake = FakeAPI([
            response(calls=[call("read", {"file": "subject.py"})], finish="tool_calls"),
            response(calls=[call("write", {"file": "subject.py", "source": CASES["fallback"]["reference"]})], finish="tool_calls"),
            response(content="Implemented cap; historical rationale unavailable."),
        ])
        result = self.run_fake(fake)
        self.assertEqual(result["termination"], "completed")
        self.assertTrue(result["grade"]["task_success"])
        self.assertIsNone(result["grade"]["overall_success"])
        self.assertIn("+def retry_delay(seconds, cap=30):", result["patch"])
        initial = json.dumps(fake.payloads[0])
        self.assertNotIn("Nonpositive seconds disables", initial)
        self.assertNotIn("rubric", initial)
        self.assertEqual(len(result["usage"]), 3)

    def test_history_tool_in_a_is_denied_even_if_model_invents_it(self):
        fake = FakeAPI([response(calls=[call("history", {})], finish="tool_calls"), response()])
        result = self.run_fake(fake)
        self.assertIn("not permitted", result["events"][0]["error"])

    def test_tool_argument_cannot_override_dispatch(self):
        fake = FakeAPI([response(calls=[call("read", {"tool": "history"})], finish="tool_calls"), response()])
        self.assertIn("Invalid tool", self.run_fake(fake)["events"][0]["error"])

    def test_429_no_retry_persists_failed_request(self):
        fake = FakeAPI([RunError("HTTP 429; no automatic retry")])
        result = self.run_fake(fake)
        self.assertEqual(result["termination"], "infrastructure_error")
        self.assertEqual(len(result["requests"]), 1)
        self.assertEqual(result["responses"], [])
        self.assertFalse(result["grade"]["task_success"])

    def test_429_retries_are_bounded_logged_and_preserve_model(self):
        settings = dict(self.settings(), max_retries=2, max_retry_wait=30, max_http_requests=3)
        fake = FakeAPI([RateLimited(0), response()])
        with tempfile.TemporaryDirectory() as temp:
            result = run_attempt({"model": "test/model:free", "case": "fallback", "condition": "A"},
                settings, Path(temp), "test-secret", api=fake, sleep=lambda seconds: None)
        self.assertEqual(result["termination"], "completed")
        self.assertEqual(len(result["requests"]), 2)
        self.assertEqual(len(result["rate_limit_events"]), 1)
        self.assertEqual(fake.payloads[0], fake.payloads[1])
        with tempfile.TemporaryDirectory() as temp:
            failed = run_attempt({"model": "test/model:free", "case": "fallback", "condition": "A"},
                settings, Path(temp), "test-secret", api=FakeAPI([RateLimited(3600)]), sleep=lambda s: self.fail("Must not wait an hour"))
        self.assertEqual(failed["termination"], "infrastructure_error")
        self.assertEqual(len(failed["requests"]), 1)

    def test_retry_after_seconds_and_dates(self):
        self.assertEqual(retry_seconds("12"), 12)
        self.assertIsNone(retry_seconds("invalid"))
        self.assertIsNone(retry_seconds("nan"))
        self.assertIsNone(retry_seconds("inf"))
        self.assertEqual(retry_seconds("Wed, 01 Jan 2020 00:00:00 GMT"), 0)

    def test_truncated_answer_is_not_scored_as_success(self):
        result = self.run_fake(FakeAPI([response(finish="length")]))
        self.assertEqual(result["termination"], "incomplete_response")

    def test_model_change_and_nonzero_cost_stop(self):
        for extra, reason in [({"model": "different/model"}, "model_mismatch"),
                              ({"usage": {"cost": 0.01}}, "nonzero_reported_cost")]:
            result = self.run_fake(FakeAPI([response(**extra)]))
            self.assertEqual(result["termination"], reason)

    def test_turn_budget_terminates(self):
        settings = dict(self.settings(), max_turns=1)
        result = self.run_fake(FakeAPI([response(calls=[call("read", {"file": "subject.py"})], finish="tool_calls")]), settings=settings)
        self.assertEqual(result["termination"], "turn_budget")

    def test_http_error_never_reads_or_exposes_body(self):
        class Opener:
            def open(self, request, timeout):
                raise urllib.error.HTTPError(request.full_url, 401, "secret-body", {}, None)
        with self.assertRaisesRegex(RunError, "HTTP 401") as caught:
            API("private-key", opener=Opener()).request("key")
        self.assertNotIn("private-key", str(caught.exception))
        self.assertNotIn("secret-body", str(caught.exception))

    def test_summary_preserves_failures_unstarted_rows_and_missing_usage(self):
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp)
            plan = make_plan(["test/model:free"], "smoke", 42)
            save(directory / "manifest.json", {"plan": plan, "phase": "smoke", "status": "stopped_on_error"})
            attempt = directory / plan[0]["id"]
            attempt.mkdir()
            save(attempt / "attempt.json", {"termination": "infrastructure_error", "usage": [None],
                                            "requests": [{}], "grade": {"task_success": False}})
            result = summarize(directory)
            self.assertEqual(result["planned"], 3)
            self.assertEqual(result["attempted"], 1)
            self.assertIsNone(result["rows"][0]["prompt_tokens"])
            self.assertEqual(sum(r["termination"] == "not_started" for r in result["rows"]), 2)

    def test_parent_watchdog_preserves_journal_and_stops_at_request_cap(self):
        class CatalogAPI:
            def __init__(self, key):
                pass
            def request(self, *args, **kwargs):
                if args[0] == "key":
                    return {"data": {"free_model_daily_requests": {"remaining": 50, "limit": 50}}}
                return {"data": [{"id": "test/model:free", "pricing": {"prompt": "0", "completion": "0"},
                                  "supported_parameters": ["tools", "max_tokens", "temperature"]}]}
        class Worker:
            def __init__(self, target, args):
                self.spec, self.settings, self.directory, self.key = args
                self.alive = True
            def start(self):
                save(self.directory / "attempt.json", {"spec": self.spec, "termination": "running", "requests": [{}]})
            def join(self, timeout=None):
                pass
            def is_alive(self):
                return self.alive
            def terminate(self):
                self.alive = False
        with tempfile.TemporaryDirectory() as temp, patch.dict(os.environ, {"OPENROUTER_API_KEY": "test-secret"}):
            directory = Path(temp)
            binary = directory / "binary"
            binary.write_bytes(b"fake executable never launched")
            output = directory / "results"
            args = SimpleNamespace(binary=str(binary), max_turns=2, max_tokens=100,
                seconds=1, interval=0, phase="smoke", seed=42, execute=True,
                max_requests=6, max_retries=0, output=output)
            with patch("openrouter_runner.API", CatalogAPI), patch("openrouter_runner.multiprocessing.get_context", return_value=SimpleNamespace(Process=Worker)):
                execute(args, ["test/model:free"])
            manifest = json.loads((output / "manifest.json").read_text())
            self.assertEqual(manifest["status"], "finished_requires_human_review")
            self.assertEqual(manifest["requests_used"], 3)
            attempt = json.loads((output / "attempt-001" / "attempt.json").read_text())
            self.assertEqual(attempt["termination"], "deadline_or_worker_failure")
            self.assertFalse(attempt["grade"]["task_success"])

    def test_no_attempt_starts_without_full_comparison_budget(self):
        class Catalog:
            def __init__(self, key):
                pass
            def request(self, endpoint, **kwargs):
                if endpoint == "key":
                    return {"data": {"free_model_daily_requests": {"remaining": 50, "limit": 50}}}
                return {"data": [{"id": "test/model:free", "pricing": {"prompt": "0", "completion": "0"},
                                  "supported_parameters": ["tools", "max_tokens", "temperature"]}]}
        with tempfile.TemporaryDirectory() as temp, patch.dict(os.environ, {"OPENROUTER_API_KEY": "admission-test-key"}):
            directory = Path(temp)
            binary = directory / "binary"
            binary.write_bytes(b"fake")
            args = SimpleNamespace(binary=str(binary), max_turns=8, max_tokens=100,
                seconds=1, interval=6, phase="smoke", seed=42, execute=True,
                max_requests=29, max_retries=2, output=directory / "results")
            with patch("openrouter_runner.API", Catalog), patch("openrouter_runner.multiprocessing.get_context") as context:
                execute(args, ["test/model:free"])
                context.return_value.Process.assert_not_called()
            manifest = json.loads((args.output / "manifest.json").read_text())
            self.assertEqual(manifest["status"], "insufficient_block_budget")
            self.assertEqual(manifest["requests_used"], 0)
            self.assertEqual(len(manifest["results"]), 0)
            # Safe resume keeps the plan, and still refuses to shrink budgets.
            args.resume = True
            with patch("openrouter_runner.API", Catalog):
                execute(args, ["test/model:free"])
            resumed = json.loads((args.output / "manifest.json").read_text())
            self.assertEqual(manifest["plan"], resumed["plan"])
            self.assertEqual(len(resumed["invocations"]), 2)
            binary.write_bytes(b"changed")
            with self.assertRaisesRegex(RunError, "changed"):
                execute(args, ["test/model:free"])


if __name__ == "__main__":
    unittest.main()
