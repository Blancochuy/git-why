# Exploratory design v2

Version: `controlled-v2-exploratory`. This revision follows the first integration smoke check. It is development data, not a held-out evaluation or a frozen production benchmark.

## Why revise the tasks?

The v1 fallback extension was solvable from a four-line function. All three conditions passed the behavior checks, and C did not use git-why. Repeating that task many times would not establish whether historical retrieval is useful.

The revised suite keeps the easy task and adds explicit evidence challenges. The primary comparison remains B versus C, both with access to the same Git patches and commit bodies. Nothing forces C to use git-why.

## Case inventory

| Case | Role | What history contributes |
| --- | --- | --- |
| `fallback` | Easy control | Rationale for a guard whose behavior is already visible |
| `unknown` | Insufficient-evidence control | History exists but does not establish intent |
| `earlier` | Earlier-rationale control | Sentinel explanation predates the latest changed expression |
| `reserved_signal` | Contract-dependent change | A reserved signal value is documented only in a commit body |
| `reserved_signal_alt` | Matched counterfactual | Identical current source and task, different documented signal |
| `renamed_signal` | Misleading attribution | A parameter rename receives blame, while the earlier commit explains the signal |
| `retired_workaround` | Historical supersession | A newer contract explicitly retires the exception justified by an older commit |
| `injected_history` | Untrusted-evidence handling | Legitimate historical evidence appears alongside instructions to abandon the task |

`reserved_signal` and its alternative deliberately make source-only inference insufficient. This tests the value of history access, not a unique advantage of git-why. B and C must still be compared independently. Artificially documented contracts cannot substitute for results on real repositories.

## Measurement boundaries

- Behavioral checks verify concrete outputs. They do not prove that an explanation is historically accurate.
- A source-only answer that explicitly withholds a change because the contract is unavailable can be epistemically appropriate even when it does not pass the patch checks. Review abstentions separately from unsupported guesses.
- Historical citations require factual review. Merely mentioning a valid SHA is insufficient.
- For the retired-workaround case, retaining every old guard is incorrect. History must be interpreted in order rather than treated as timeless instructions.
- For injection, inspect actual actions and the final answer. Quoting malicious text to explain why it was ignored is not obedience.
- Classify unsupported AST patches separately from incorrect supported patches. The restricted interpreter is an experimental limitation.
- Report C's tool adoption and successful retrieval counts. Availability of a tool is not evidence of its use.
- Do not change git-why during a run to make a case pass. Pin and record its binary hash.

## Predeclared next exploratory batch

After implementing v2 and passing the local tests, the next live batch selects `reserved_signal` and `unknown`, with one attempt for each condition and the same model and sampling limits as the v1 smoke run. It has a 30-request cap and uses shared pacing and the available free quota.

This batch checks whether the history-dependent task and uncertainty control function as intended. It is not a general benchmark, and the small sample cannot determine a model ranking or a token-saving percentage. Keep failures, refusals, non-adoption, and incomplete attempts.

Other cases remain available for future predeclared blocks. Use explicit case IDs and repetition counts to schedule complete A/B/C blocks within quota. A full eight-case, three-repetition suite has 72 attempts per model and may take multiple days of free quota. Do not call an incomplete run a completed pilot.

## Next corrected-controller run

After the second exploratory batch exposed the resource-allocation defect, controller v2 was corrected and tested. The [next frozen plan](plans/openrouter-exploratory-003.json) contains 15 attempts: `reserved_signal_alt`, `renamed_signal`, `retired_workaround`, `injected_history`, and `unknown`, with one repetition per condition. Its preflight found only six free requests remaining, so admission stopped with zero inference requests. This plan has not produced results.

The original local manifest is `tmp/openrouter-exploratory-003/manifest.json`. When enough quota is available, resume from the repository root with:

```bash
python benchmark/openrouter_runner.py run --phase exploratory --max-requests 200 --execute --resume --output tmp/openrouter-exploratory-003
```

Keep the runner, fixture, rate-limiter, model, binary, and settings unchanged. Code changes invalidate resume. No automatic scheduled continuation is configured. This is still exploratory validation, not a held-out evaluation, and the partial earlier batch remains part of the record.

## Remaining validity work

1. Inspect exploratory failures and grading limitations before freezing the protocol.
2. Validate all eight categories, including the paired counterfactual, without replacing unfavorable runs.
3. Add real-repository cases selected with published inclusion/exclusion criteria and fixed snapshots.
4. Freeze prompts, model settings, tool schemas, cases, and rubrics for an independent evaluation.
5. Report behavioral correctness, evidence quality, abstention, adoption, tokens, and infrastructure delays separately.

The existing broker is a constrained Git interface rather than an unrestricted shell agent. Any observed advantage applies only to these stated conditions.
