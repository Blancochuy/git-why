# Exploratory history-contract check

Date: 2026-10-21. Model: `cohere/north-mini-code:free`. Design: `controlled-v2-exploratory`, controller v1. This is development evidence, not a completed pilot or a model ranking.

## Plan and limits

Selected before execution: `reserved_signal` and `unknown`, one repetition of A/B/C per task. Temperature 0, 2,048 output tokens per turn, nominal eight turns per attempt, 180-second deadline, 30 HTTP requests across the batch. No git-why change or model switch was made during the run.

The three reserved-signal attempts all received eight turns. The controller reduced the later unknown attempts to seven turns for C and two for B as the request cap approached; A never started. This is a discovered resource-allocation defect. The unknown block is not a valid matched comparison. It must remain visible and must not be silently replaced.

## Reserved-signal block

The source clamps all negative inputs. The task asks for signed offsets while preserving the reserved wire signal; its exact value is documented only in a commit body.

| Condition | Final patch checks | Final answer | HTTP requests | git-why calls | Prompt tokens | Completion tokens |
| --- | --- | --- | --- | --- | --- | --- |
| A: source | Failed | No; eight turns exhausted | 8 | Unavailable | 3,600 | 478 |
| B: Git | Passed | Yes | 7 | Unavailable | 5,472 | 1,366 |
| C: Git + git-why | Passed | No; eight turns exhausted | 8 | 3 | 5,732 | 826 |

B cited the commit documenting the reserved signal and described the corresponding contract in its final answer. A repeatedly read the source and did not produce a change. C used read, history, blame, three context calls, another read, and write. Its final patch passed all five behavior checks, but it had no remaining turn for a final explanation.

The original controller's summary forced `task_success=false` for incomplete attempts. That field conflated completion with behavioral correctness. The table above derives patch correctness directly from the recorded individual checks and reports completion separately. New controller summaries separate these fields without modifying the original attempt journals.

This single block shows that history access can supply a contract absent from source. It does not show a git-why advantage over Git. C did use git-why, unlike the first smoke run, but it needed more calls and did not finish within the same turn limit. C's lower completion-token count is not an efficiency win because its answer was incomplete.

## Insufficient-evidence block

- C completed after five requests, using read, history, and two Git show calls. It did not call git-why.
- Its final answer described the code and evolution, but did not explicitly acknowledge that the underlying motive was not documented. This is a qualitative inspection note, not a completed blinded human grade.
- B received only two turns and exhausted them. A was not started.
- No B/C quality or efficiency comparison is valid for this block because the configured turn limits differed.

## Infrastructure and cost

The batch recorded 30 inference requests. Every returned usage record reported zero cost. No HTTP 429 was observed. Pacing waits remain in the journals; wall-clock time is not used here to claim model or retrieval speed.

## Changes justified by this run

Controller v2 admits an entire A/B/C block only when the global request cap, reported quota, and local rolling quota can cover its declared maximum. It never reduces turn limits to fit remaining requests. Each attempt gets a fixed HTTP allowance of `max_turns + max_retries` shared across its turns; 429 retries remain bounded per request as well.

The controller can explicitly resume only unstarted work in an unchanged frozen run. It verifies runner/fixture/rate-limiter hashes, binary hash, model selection, sampling limits, retry policy, total request cap, and phase. It never regenerates an already started attempt. Legacy controller-v1 runs, including this one, cannot be resumed.

The next blocks must be recorded separately under the corrected controller. The larger suite and real-repository cases are still pending. Do not tune a larger turn limit only for C; any new budget must apply equally to all conditions and be declared before running them.

## Audit files

Raw journals and the original controller summary remain unchanged under the ignored local directory `tmp/openrouter-exploratory-002/`. Each attempt records the actual effective settings, requests/responses, tool events, final source, checks, and termination reason.
