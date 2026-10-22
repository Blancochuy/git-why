"""Offline harness validation, not a model benchmark. Python standard library only."""

import argparse
import ast
import hashlib
import json
import operator
import os
from pathlib import Path
import random
import re
import subprocess
import tempfile
import time


VERSION = "controlled-v1"
CASES = {
    "fallback": {
        "prompt": "Extend retry_delay(seconds) with an optional cap=30 parameter. Preserve existing behavior for current callers and explain constraints relevant to your change.",
        "history": [
            ("Add retry scheduling", "def retry_delay(seconds):\n    return min(seconds, 30)\n"),
            ("Preserve disabled retries\n\nNonpositive seconds disables retries. The dispatcher expects exactly zero, never a negative delay.",
             "def retry_delay(seconds):\n    if seconds <= 0:\n        return 0\n    return min(seconds, 30)\n"),
        ],
        "reference": "def retry_delay(seconds, cap=30):\n    if seconds <= 0:\n        return 0\n    return min(seconds, cap)\n",
        "mutant": "def retry_delay(seconds, cap=30):\n    return min(seconds, cap)\n",
        "checks": [("retry_delay", [5], 5), ("retry_delay", [50], 30),
                   ("retry_delay", [50, 10], 10), ("retry_delay", [-5], 0),
                   ("retry_delay", [0], 0)],
        "rubric": "Explain nonpositive input means disabled retries, cite the supporting commit, and do not invent an incident.",
    },
    "unknown": {
        "prompt": "Explain why normalize(value) returns None for an empty string. Separate observable behavior from historical intent. Do not change the code.",
        "history": [
            ("Initial implementation", "def normalize(value):\n    return value\n"),
            ("Update normalization", "def normalize(value):\n    if value == '':\n        return None\n    return value\n"),
        ],
        "rubric": "Describe the empty-string special case, acknowledge history does not establish why, and reject invented motivations. A citation alone does not prove intent.",
    },
    "earlier": {
        "prompt": "Extend timeout(seconds) with an optional multiplier=1 parameter for scaling nonnegative values. Preserve compatibility and explain relevant constraints.",
        "history": [
            ("Add timeout conversion", "def timeout(seconds):\n    return seconds * 1000\n"),
            ("Preserve unlimited timeout sentinel\n\nNegative seconds means unlimited. Downstream adapters require the sentinel -1, not a scaled negative value.",
             "def timeout(seconds):\n    if seconds < 0:\n        return -1\n    return seconds * 1000\n"),
            ("Clarify numeric literal", "def timeout(seconds):\n    if seconds < 0:\n        return -1\n    return seconds * 1_000\n"),
        ],
        "reference": "def timeout(seconds, multiplier=1):\n    if seconds < 0:\n        return -1\n    return seconds * 1_000 * multiplier\n",
        "mutant": "def timeout(seconds, multiplier=1):\n    return seconds * 1_000 * multiplier\n",
        "checks": [("timeout", [2], 2000), ("timeout", [2, 3], 6000),
                   ("timeout", [-2], -1), ("timeout", [-2, 3], -1),
                   ("timeout", [0, 3], 0)],
        "rubric": "Find and cite the earlier sentinel rationale, preserve -1 for negative input, and distinguish the later numeric-literal edit from the reason for the guard.",
    },
}


def git_env():
    # Avoid user hooks, signing, identity, remotes, and inherited injected config.
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env.update(GIT_CONFIG_NOSYSTEM="1", GIT_CONFIG_GLOBAL=os.devnull,
               GIT_AUTHOR_NAME="Fixture Author", GIT_AUTHOR_EMAIL="fixture@example.invalid",
               GIT_COMMITTER_NAME="Fixture Author", GIT_COMMITTER_EMAIL="fixture@example.invalid",
               GIT_TERMINAL_PROMPT="0", GIT_PAGER="cat")
    return env


def git(repo, *args, date=None):
    env = git_env()
    if date:
        env.update(GIT_AUTHOR_DATE=date, GIT_COMMITTER_DATE=date)
    return subprocess.run(
        ["git", "-c", "core.hooksPath=" + os.devnull, "-c", "commit.gpgsign=false",
         "-c", "core.autocrlf=false", "-c", "safe.directory=" + repo.as_posix(), *args],
        cwd=repo, env=env, capture_output=True, text=True, encoding="utf-8",
        check=True, timeout=10,
    ).stdout.strip()


def fixture(parent, case_id, condition):
    if condition not in "ABC" or len(condition) != 1:
        raise ValueError("Invalid condition")
    repo = parent / "workspace"
    repo.mkdir()
    if condition == "A":
        git(repo, "init", "-b", "main", "--separate-git-dir=" + str(parent / "private-history"))
    else:
        git(repo, "init", "-b", "main")
    shas = []
    for index, (message, source) in enumerate(CASES[case_id]["history"]):
        (repo / "subject.py").write_text(source, encoding="utf-8", newline="\n")
        git(repo, "add", "subject.py")
        git(repo, "commit", "-m", message, date=f"2025-01-0{index + 1}T12:00:00+00:00")
        shas.append(git(repo, "rev-parse", "HEAD"))
    # Condition A has no history at all, including through the broker.
    if condition == "A":
        (repo / ".git").unlink()
    return repo, shas


class Broker:
    """Only this interface is intended for a future tool-only model adapter.

    No shell, arbitrary paths, test execution, URLs, or unrestricted Git args.
    This is an application capability boundary, not an OS sandbox.
    """

    def __init__(self, repo, condition, max_calls=20, deadline_seconds=30, binary=None):
        self.repo, self.condition = repo, condition
        self.max_calls = max_calls
        self.started = time.monotonic()
        self.deadline_seconds = deadline_seconds
        self.events = []
        self.binary = Path(binary).resolve() if binary else None

    def call(self, tool, **args):
        started = time.monotonic()
        event = {"tool": tool, "arguments": args}
        try:
            if len(self.events) >= self.max_calls:
                raise ValueError("Tool call budget exhausted")
            if started - self.started >= self.deadline_seconds:
                raise ValueError("Attempt deadline exceeded")
            if tool == "read" and args == {"file": "subject.py"}:
                result = (self.repo / "subject.py").read_text(encoding="utf-8")
            elif tool == "write" and set(args) == {"file", "source"} and args["file"] == "subject.py":
                source = args["source"]
                if not isinstance(source, str) or len(source.encode()) > 16000:
                    raise ValueError("Invalid source size")
                ast.parse(source)
                (self.repo / "subject.py").write_text(source, encoding="utf-8", newline="\n")
                result = "Saved subject.py"
            elif tool == "history" and self.condition in "BC" and args == {}:
                result = git(self.repo, "log", "--format=%H%n%B", "--", "subject.py")
            elif tool == "blame" and self.condition in "BC" and args == {}:
                result = git(self.repo, "blame", "--porcelain", "--", "subject.py")
            elif tool == "show" and self.condition in "BC" and set(args) == {"sha"} and isinstance(args["sha"], str) and re.fullmatch(r"[0-9a-f]{40}", args["sha"]):
                if args["sha"] not in git(self.repo, "rev-list", "HEAD").splitlines():
                    raise ValueError("Commit is not in the permitted history")
                result = git(self.repo, "show", "--no-ext-diff", "--no-textconv", "--format=fuller", args["sha"], "--", "subject.py")
            elif tool == "context" and self.condition == "C" and self.binary and set(args) == {"file", "line"} and args["file"] == "subject.py" and type(args["line"]) is int and 1 <= args["line"] <= 1000:
                requests = [
                    {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}},
                    {"jsonrpc": "2.0", "method": "notifications/initialized"},
                    {"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {
                        "name": "git-why-context", "arguments": args}},
                ]
                # Fresh process per call, no remotes, no inherited provider tokens.
                env = git_env()
                env = {k: v for k, v in env.items() if not any(s in k.upper() for s in ("TOKEN", "SECRET", "API_KEY"))}
                env.update(GIT_CONFIG_COUNT="1", GIT_CONFIG_KEY_0="safe.directory",
                           GIT_CONFIG_VALUE_0=self.repo.as_posix())
                response = subprocess.run([str(self.binary), "mcp"], cwd=self.repo, env=env,
                    input="".join(json.dumps(r) + "\n" for r in requests), capture_output=True,
                    text=True, encoding="utf-8", timeout=10, check=True)
                messages = [json.loads(line) for line in response.stdout.splitlines()]
                message = next(m for m in messages if m.get("id") == 2)
                if "error" in message or message["result"].get("isError"):
                    raise ValueError("MCP tool returned an error")
                result = json.dumps(message["result"])
            else:
                raise ValueError("Operation or arguments not permitted")
            # Explicitly log truncation, never execute model-generated code.
            event.update(result=result[:24000], truncated=len(result) > 24000, error=None)
        except (ValueError, SyntaxError, OSError, StopIteration, KeyError, subprocess.SubprocessError) as error:
            event.update(result=None, error=str(error))
        event["elapsed_seconds"] = time.monotonic() - started
        self.events.append(event)
        return event


def evaluate(source, name, values):
    """Evaluate a tiny pure-function subset without exec, imports, or attribute access."""
    if len(source.encode()) > 16000:
        raise ValueError("Source too large")
    tree = ast.parse(source)
    if sum(1 for _ in ast.walk(tree)) > 200:
        raise ValueError("Expression budget exceeded")
    if len(tree.body) != 1 or not isinstance(tree.body[0], ast.FunctionDef):
        raise ValueError("Expected exactly one function")
    fn = tree.body[0]
    if fn.name != name or fn.decorator_list or fn.args.vararg or fn.args.kwarg or fn.args.kwonlyargs or fn.args.posonlyargs:
        raise ValueError("Unsupported function")
    names = [a.arg for a in fn.args.args]
    if len(values) > len(names) or len(values) < len(names) - len(fn.args.defaults):
        raise ValueError("Incorrect arguments")
    scope = {}
    ops = {ast.Add: operator.add, ast.Sub: operator.sub, ast.Mult: operator.mul,
           ast.Lt: operator.lt, ast.LtE: operator.le, ast.Eq: operator.eq}

    def expr(node):
        if isinstance(node, ast.Constant) and (node.value is None or type(node.value) in (int, str)):
            return node.value
        if isinstance(node, ast.Name):
            return scope[node.id]
        if isinstance(node, ast.UnaryOp) and isinstance(node.op, ast.USub):
            return -expr(node.operand)
        if isinstance(node, ast.BinOp) and type(node.op) in ops:
            left, right = expr(node.left), expr(node.right)
            if type(left) is not int or type(right) is not int or max(abs(left), abs(right)) > 10**12:
                raise ValueError("Arithmetic outside fixture limits")
            return ops[type(node.op)](left, right)
        if isinstance(node, ast.Compare) and len(node.ops) == 1 and type(node.ops[0]) in ops:
            return ops[type(node.ops[0])](expr(node.left), expr(node.comparators[0]))
        if isinstance(node, ast.Call) and isinstance(node.func, ast.Name) and node.func.id in ("min", "max") and not node.keywords:
            return {"min": min, "max": max}[node.func.id](*[expr(a) for a in node.args])
        raise ValueError("Unsupported expression")

    for index, arg in enumerate(names):
        scope[arg] = values[index] if index < len(values) else expr(fn.args.defaults[index - (len(names) - len(fn.args.defaults))])

    def statements(nodes):
        for node in nodes:
            if isinstance(node, ast.Return):
                return True, expr(node.value)
            if isinstance(node, ast.If):
                returned, value = statements(node.body if expr(node.test) else node.orelse)
                if returned:
                    return True, value
            else:
                raise ValueError("Unsupported statement")
        return False, None

    return statements(fn.body)[1]


def grade(case_id, source):
    case = CASES[case_id]
    if "checks" not in case:
        return {"status": "human_review_required", "task_success": None, "rubric": case["rubric"]}
    try:
        outcomes = [evaluate(source, name, values) == expected for name, values, expected in case["checks"]]
        return {"status": "scored", "task_success": all(outcomes), "checks": outcomes,
                "historical_accuracy": None, "review_required": True}
    except (ValueError, SyntaxError, KeyError, TypeError, RecursionError) as error:
        return {"status": "unsupported_or_invalid_patch", "task_success": False, "error": str(error)}


def dry_run(destination, seed=42):
    """Omniscient reference driver validates wiring only; it is NOT an agent."""
    destination.mkdir(parents=True, exist_ok=False)
    rng = random.Random(seed)
    attempts = []
    for case_id in CASES:
        order = list("ABC")
        rng.shuffle(order)
        for condition in order:
            started = time.monotonic()
            with tempfile.TemporaryDirectory(prefix="gitwhy-fixture-") as temp:
                repo, shas = fixture(Path(temp), case_id, condition)
                broker = Broker(repo, condition)
                broker.call("read", file="subject.py")
                if condition in "BC":
                    broker.call("history")
                case = CASES[case_id]
                if "reference" in case:
                    broker.call("write", file="subject.py", source=case["reference"])
                source = (repo / "subject.py").read_text(encoding="utf-8")
                attempts.append({"case": case_id, "condition": condition, "source_sha": shas[-1],
                                 "prompt": case["prompt"], "events": broker.events,
                                 "grade": grade(case_id, source), "final_source": source,
                                 "elapsed_seconds": time.monotonic() - started,
                                 "termination": "reference_driver_completed", "usage": None})
    result = {"protocol": VERSION, "kind": "harness_self_check_not_model_evaluation",
              "seed": seed, "model": None, "mcp_available": False,
              "warning": "Reference driver knows the solutions and does not exercise MCP. No effectiveness conclusions are valid.",
              "harness_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
              "attempts": attempts}
    (destination / "self-check.json").write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    return result


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--self-check", type=Path, required=True, help="New output directory; never overwrites a run")
    parser.add_argument("--seed", type=int, default=42)
    args = parser.parse_args()
    result = dry_run(args.self_check, args.seed)
    print(f"Validated {len(result['attempts'])} fixture/condition combinations. NOT model results.")
