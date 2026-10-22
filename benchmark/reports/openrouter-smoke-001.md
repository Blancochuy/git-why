# OpenRouter integration smoke check

Date: 2026-10-21. Phase: integration smoke test, not an effectiveness benchmark.

## Configuration

- Model requested: `cohere/north-mini-code:free`.
- Task: controlled `fallback` fixture, one attempt per condition.
- Order: B, A, C, randomized with seed 42.
- Temperature: 0. Output cap: 2,048 tokens per turn.
- Limits: 8 turns and 20 tool calls per attempt, 180-second attempt deadline, 18 HTTP requests across the run.
- Pacing: at least six seconds between request starts, shared across processes. Up to two bounded 429 retries; none were needed.
- Each attempt had a fresh workspace and conversation. No answer, private check, or expected evidence was sent to the model.

## Observations

| Condition | Behavior checks | HTTP requests | Tool calls | git-why calls | Prompt tokens | Completion tokens | Elapsed seconds | Pacing wait seconds |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| A: source only | Passed | 3 | 2 | Unavailable | 941 | 1,101 | 23.06 | 7.54 |
| B: Git | Passed | 7 | 6 | Unavailable | 5,176 | 1,261 | 40.45 | 26.51 |
| C: Git + git-why | Passed | 4 | 3 | 0 | 2,069 | 830 | 24.25 | 12.46 |

All attempts reached a final answer. The 14 inference responses reported zero cost. No HTTP 429 occurred. Token counts are provider-reported sums over successful responses, including repeated conversation context. Elapsed time includes pacing waits and is not a pure model or tool-latency measurement.

The initial account endpoint reported 50 free requests remaining. A post-run account query reported 40 remaining. The provider's account counter did not match the runner's 14 recorded responses at the time of observation; the cause was not established. Keep both observations rather than reconciling them by assumption. Local pacing counts all 14 request reservations conservatively.

## Interpretation

This run validates API authentication, tool dispatch, isolated conversations, patch capture, usage capture, and post-session behavioral checks. It does not establish overall task success: historical-accuracy and unsupported-claim review remain pending.

C did not call git-why. Its result therefore cannot be attributed to retrieving evidence with git-why. A also passed, indicating that this small modification can be solved from the current source alone. This is a ceiling-effect warning for the proposed quality comparison, not evidence that history never helps.

Do not force C to call git-why or replace this outcome with a more favorable rerun. Any future forced-retrieval experiment must be a separately labeled condition. Validate the other fixture categories, add cases whose requested changes meaningfully depend on historical constraints, and freeze the revised design before evaluation.

## Audit data

Full request/response transcripts, selected catalog metadata, source and binary hashes, tool events, final source, patches, and grading output remain locally in `tmp/openrouter-smoke-001/`. That ignored directory contains `manifest.json`, per-attempt journals, `summary.json`, and a human-review packet. It is not committed. This report is a compact observation log, not a substitute for raw data when publishing benchmark conclusions.
