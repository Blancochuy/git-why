# AI/MCP Benchmark TODO

This document defines the benchmark plan for evaluating whether `git-why` improves LLM performance on real code-understanding tasks.

## Goal

Measure the effect of `git-why` context on model quality, speed, and reliability.

Primary comparison:

- Baseline: model solves task without `git-why`
- Assisted: model solves same task with `git-why` via MCP

## TODO Checklist

- [ ] Select benchmark repositories and freeze commits for reproducibility.
- [ ] Define 30-100 tasks focused on "why does this code exist?" and safe code changes.
- [ ] Build an evaluation harness that runs each task in Baseline and Assisted modes.
- [ ] Add support for multiple models/providers.
- [ ] Add automatic scoring plus human spot-check review.
- [ ] Publish benchmark scripts, prompts, and raw results in this repo.

## Candidate Model Set

Use at least one model from each tier:

- Frontier large model
- Cost-efficient mid model
- Open-weight local model

Example groups (update over time):

- OpenAI: GPT-5 class and a smaller model
- Anthropic: Claude class model
- Google: Gemini class model
- Open-weight: Qwen/Llama class model

## Task Design

Use tasks that require historical intent, not just static code reading.

Task categories:

1. Explain intent of a suspicious line or condition.
2. Link a behavior to likely historical reason from commit metadata.
3. Propose a minimal fix while preserving original intent.
4. Reject a misleading change request when history indicates risk.

Each task should include:

- repository + commit SHA
- target file and line/range/function
- user prompt
- reference answer/rubric
- expected risk constraints

## Suggested Standard Benchmarks to Reuse

Reuse common coding/evaluation practices where possible:

- SWE-bench style issue-resolution flow (adapted for historical-context tasks)
- HumanEval-like pass/fail checks for small patch tasks
- Repo-specific golden answers for explanation quality

## Evaluation Metrics

Track both quality and cost:

- Task success rate (% passing rubric)
- Factual accuracy about history (commit/author/date/message correctness)
- Hallucination rate (% claims unsupported by repo history)
- Patch correctness (tests pass + no forbidden changes)
- Time-to-first-valid-answer
- Token usage and estimated cost

## Experiment Protocol

- Run each model on the same tasks with fixed seeds/settings.
- Keep temperature and max tokens consistent across Baseline and Assisted modes.
- Use identical system prompts except for MCP tool availability.
- Repeat runs (at least 3) and report mean + variance.

## Reporting Format

Publish results in a table like:

| Model | Mode | Success % | Hallucination % | Avg Time (s) | Avg Cost |
|------|------|-----------|-----------------|--------------|----------|
| Model A | Baseline | ... | ... | ... | ... |
| Model A | Assisted (`git-why`) | ... | ... | ... | ... |

Also include:

- failure examples
- where `git-why` helped most
- where `git-why` did not help

## Definition of Done

Benchmark is complete when:

- at least 30 tasks are stable and reproducible
- at least 4 models are evaluated
- scripts/prompts/results are public in this repo
- a short methodology + findings report is published
