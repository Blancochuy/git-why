# Agent history benchmark

Status: offline fixtures, restricted tool broker, reference-driver self-check, and tests are implemented. No model runs or results have been produced. A provider adapter and complete attempt orchestration are still pending.

## Run the harness checks

Requires Python 3.10+ and Git. No third-party Python packages or API keys are needed.

```bash
python -m unittest discover -s benchmark -p 'test_*.py' -v
python benchmark/harness.py --self-check tmp/benchmark-check-001
```

Each self-check requires a new output directory and writes `self-check.json`. It records nine fixture/condition combinations using a reference driver that already knows the solutions. These are wiring checks, not model attempts, and cannot measure effectiveness.

Set `GITWHY_BENCH_BINARY` to the absolute path of a local git-why binary to enable the real MCP integration test. Without it that test is explicitly skipped. CI builds the binary and runs all tests on Ubuntu and Windows.

### Implemented boundaries and remaining work

- The broker permits only `subject.py` reads/writes and fixed Git history/blame operations. C additionally supports a validated single-line MCP call when a binary is supplied.
- No shell, arbitrary Git arguments, URL tool, or Python execution tool is exposed. The current broker is intentionally narrower than a full coding agent's environment; report this limitation when comparing tools.
- Private solutions and checks remain in the controller. A's visible workspace contains no Git metadata; its fixture-building history stays outside the exposed workspace. All conditions receive identical source bytes.
- Patch checks use a small AST interpreter, with no `exec`, imports, decorators, attribute access, loops, or user-defined calls. Unsupported patches are classified explicitly. This limits the implementation styles accepted by the pilot; broader grading needs a separate sandbox design.
- Explanation quality and historical accuracy require human review. Passing behavior checks alone is not overall success.
- This is a tool capability boundary, not an OS sandbox. A future model adapter must receive only the prompt and broker tools, never controller filesystem or shell access. Do not run a local autonomous process against these fixtures without additional isolation.
- MCP smoke tests use fresh processes and repositories without remotes. The self-check does not exercise MCP or a model. It records missing usage as null and labels itself as a harness check.
- Complete model-run deadlines, response limits, termination handling, final-answer capture, patches, tool schemas, and provider usage integration must be implemented before any paid pilot. The broker currently enforces call budgets and checks elapsed time between calls, with subprocess timeouts; this is not yet a hard end-to-end attempt deadline.

## Question

Does adding git-why help a coding agent preserve historical constraints more reliably or with less effort than ordinary Git access?

## Two stages

1. Validate the harness with controlled repositories. Construct small histories with explicit constraints, missing rationale, and relevant earlier changes. These cases check measurement and isolation; they are not evidence of general usefulness.
2. Freeze the protocol and evaluate historical tasks from existing open-source repositories. Select cases before running any evaluated condition and report every selected case, including failures.

Begin with three controlled cases, three conditions, and three repetitions: 27 attempts. Run one attempt per condition first to verify the harness. These smoke runs belong to harness development and must not be mixed into the frozen evaluation.

## Conditions

| Condition | Agent access |
| --- | --- |
| A: source | Current source snapshot and a restricted file/test interface; no history or network |
| B: Git | Same source snapshot plus history through restricted Git operations |
| C: Git + git-why | Same source and Git operations as B, plus git-why through MCP |

B versus C is the primary comparison. A measures the value of access to history, not the unique value of git-why.

Use the same model version, task text, reasoning settings, output limits, wall-clock limit, and fresh sessions. Publish tool schemas, tool descriptions, system prompts, and permissions. C may choose whether to use git-why; do not force a call or provide the answer location. Include unused-tool runs in C's results.

Randomize condition order within each task and repetition. Reset source and process state for every attempt. Apply the same cache policy to all runs and record it.

## Controlled cases

| Case | History | Task | Success criteria |
| --- | --- | --- | --- |
| Regression constraint | A commit body explains why a fallback is necessary | Simplify or extend a related function | Satisfy the requested change and retain the protected behavior |
| Insufficient evidence | Commit messages do not explain the rationale | Explain a suspicious condition | State what is observable and acknowledge that intent cannot be established |
| Earlier rationale | A later edit receives blame while an earlier commit contains the constraint | Change behavior without restoring a regression | Find the earlier evidence and pass the task and regression checks |

Keep current-source comments neutral so they do not directly reveal the historical answer. Ensure each task has a plausible source-only interpretation. Validate that private checks reject an intentionally incorrect patch and accept the reference solution. Explanation-only cases use a fixed human-review rubric instead of claiming test-based correctness.

## Selecting real cases

Prefer Python or JavaScript libraries with lightweight offline tests so the first pilot does not depend on Go being installed. Repository popularity is not a selection metric.

Before evaluation, enumerate candidate historical changes and publish inclusion and exclusion reasons. A selected case must have:

- A redistributable license and a pinned upstream URL and commit SHA.
- A concrete constraint documented in history available at the task snapshot.
- A task whose solution is not already present in accessible future history.
- A reproducible environment and checks that distinguish correct behavior from a plausible regression.
- A recorded task prompt, target snapshot, expected evidence, and frozen rubric.

Use two or more independent repositories for the first real pilot and include cases with weak or irrelevant history. Select on evidence availability and task validity, not observed model performance. State that a small, history-focused sample cannot establish broad coding performance.

## Isolation and leakage

The agent-visible workspace must contain only the permitted snapshot, history, and public tests. Reference answers, private tests, expected SHAs, grader logic, and run metadata stay outside the agent's accessible filesystem and tools.

Do not expose an unrestricted shell and call A a source-only condition. A general shell can retrieve history from another path or use network access. Use restricted operations or enforce filesystem and network isolation. B and C must not access future commits, hidden refs, remotes, or other runs.

Disable external network access for the initial comparison, including GitHub PR enrichment. Keep fixture remotes unset. For real cases, record upstream provenance outside the agent workspace and omit origin inside it. A separate experiment can measure PR enrichment later.

Only a neutral result summarizer may read private grading data after the agent attempt ends. Do not feed grader details back into a continuing attempt.

## Measures

| Measure | Definition |
| --- | --- |
| Task success | Requested behavior and historical constraint pass the private checks |
| Historical accuracy | Cited commits support the claims in the final answer |
| Unsupported claims | Assertions of historical intent without supporting accessible evidence |
| Time | End-to-end attempt duration, including tool calls |
| Tool effort | Calls by tool, tool errors, and returned text volume |
| Model usage | Provider-reported input, cached-input, output, and reasoning tokens when available |
| Adoption | Whether C calls git-why and whether those calls succeed |

Do not count text length as tokens or infer cost from missing usage fields. Record unavailable values as null. Timeouts, tool failures, and incomplete answers remain in the denominator, with separate failure categories.

Review historical accuracy using shuffled responses without condition labels where practical. Preserve the review rubric and decisions. Record possible reviewer unblinding from response content. Use deterministic private checks for patch correctness.

## Artifacts per attempt

Save task and repetition IDs, assigned condition, model identifier, settings, tool permissions, source and tool SHAs, start/end times, termination reason, raw conversation, all tool arguments/results, final answer, patch, grader output, and usage fields. Keep secrets out of transcripts and document any redactions.

Run attempts in isolated temporary workspaces. Keep private grading assets separate from public benchmark sources during execution. After all attempts finish, publish enough data and fixture-generation instructions to audit the results; avoid embedding nested Git repositories in this repository.

## Reporting and decisions

Publish per-task outcomes and counts with denominators before aggregate percentages. Three repetitions are repeated samples of the same task, not three independent tasks. Report task-level paired B/C differences and variability; do not make statistical significance claims from the tiny pilot.

- Higher success in C is preliminary quality evidence on the selected tasks.
- Similar success with lower time or reported token usage is preliminary efficiency evidence.
- Similar outcomes and effort mean the pilot did not demonstrate an advantage.
- Worse outcomes in C require inspecting failures before adding features or increasing scope.

Publish unfavorable results too. Changes to prompts, tools, cases, or grading after seeing results require a new experiment version and a fresh evaluation set.

## Implementation order

- [x] Build the three fixture histories and private reference checks.
- [x] Test broker permissions and fixture separation in A, B, and C (adapter isolation still pending).
- [ ] Implement transcript capture, limits, clean resets, and scoring output.
- [x] Run harness checks with an omniscient reference driver; do not count them as model evaluation.
- [ ] Choose a provider/model and API spending cap before paid execution.
- [ ] Smoke-test one fresh attempt per condition, then freeze the pilot protocol.
- [ ] Execute and review the 27-attempt controlled pilot.
- [ ] Freeze real-case selection, then run a separately reported real-repository pilot.

No accuracy, efficiency, or cost improvement is claimed until an evaluation is executed.
