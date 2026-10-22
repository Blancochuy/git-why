# Running with OpenRouter free models

The runner uses Python's standard library and OpenRouter's Chat Completions API. No provider SDK is required. Local tests use fake responses and never call a model.

## Local configuration

A repository-root `.env` is ignored by Git. To set up another checkout, copy [openrouter.env.example](openrouter.env.example) to `.env` and fill in:

```dotenv
OPENROUTER_API_KEY=your-local-key
OPENROUTER_MODEL=an-exact-model-id:free
GITWHY_BENCH_BINARY=/absolute/path/to/git-why
```

Do not paste the key into chats, CLI arguments, transcripts, or commits. The dotenv loader accepts literal values only and leaves existing environment variables unchanged. Relative binary paths are resolved from your current working directory. Run commands below from the repository root.

## Discover eligible models

```bash
python benchmark/openrouter_runner.py models
python benchmark/openrouter_runner.py check-key
```

`models` reads the live public catalog and lists explicit `:free` variants advertising tool support and zero advertised pricing. `check-key` makes an authenticated GET request for selected account-limit fields. Neither command generates a response.

The run preflight also requires advertised `temperature` and `max_tokens` support. Tool support in metadata is a prerequisite, not proof of working tool calls; verify with a smoke run. The catalog and eligibility can change between runs.

Use individual IDs instead of `openrouter/free` or automatic model selection. A comparison needs an identified model. Do not switch the model to recover from a failed condition.

## Preview a run without inference

```bash
python benchmark/openrouter_runner.py run --phase smoke
```

Without `--execute`, this prints the randomized plan and does not call the API. A smoke run has three attempts on the fallback case, one per condition. It validates integration and must be reported separately from evaluation.

## Execute the smoke run

```bash
python benchmark/openrouter_runner.py run --phase smoke --execute --output tmp/openrouter-smoke-001
```

Defaults: 30 generation requests per run, 8 model turns and 20 tool calls per attempt, 2,048 output tokens per turn, 180 seconds per attempt, and at least 6 seconds between request starts. The parent process enforces the attempt deadline and preserves the request journal if a worker is terminated. Git/MCP subprocesses also have their own timeouts. A response-size limit is enforced before JSON parsing.

The request cap counts API requests, including retries, not tasks. One attempt can use several requests. A 27-attempt pilot may require several days of free quota, depending on account limits and tool use. Increasing `--max-requests` does not increase OpenRouter's quota. There is no automatic resume or quota workaround. A new output directory is required for every run.

### Rate-limit policy

- Before inference, read the server-reported free quota. Stop if the API does not report usable remaining quota.
- Reserve each request atomically in `tmp/openrouter-rate-limit.sqlite`. Processes and runs in this checkout share pacing and a conservative rolling 24-hour ceiling of at most 50 local requests per key, even if the account has a larger allowance. The ledger stores only a key hash, request timestamps, and pacing state.
- Respect the remaining allowance observed at preflight. Usage by other applications or other checkouts can still cause a 429; local accounting cannot enforce a global account quota.
- Retry only HTTP 429, at most twice, with exponential backoff plus jitter. Honor `Retry-After` in seconds or HTTP-date form. Record each rejection, scheduled delay, request, and pacing wait.
- Wait at most 30 seconds for a retry and stay within the attempt deadline and request cap. A longer server hint stops the batch and leaves a shared cooldown for subsequent runs.
- Do not retry ambiguous network failures, authentication errors, or other HTTP failures. Do not switch models or silently regenerate completed attempts.
- Keep infrastructure delays and rate-limit events in reports. End-to-end elapsed time includes waiting, so it cannot alone establish a retrieval-efficiency improvement.

Use `--max-retries 0` to disable 429 retries. Use the same retry policy and interval across conditions and models. The six-second floor cannot be reduced through the CLI. Do not delete the ledger to evade the free quota.

The entire batch stops on exhausted quota, unrecovered infrastructure errors, mismatched models, or nonzero reported cost. Failed and unstarted work stay in the manifest. Correct the issue and report a new run separately; never silently omit failures or replace them with favorable reruns.

## Pilot and multiple models

After inspecting smoke transcripts and freezing the configuration:

```bash
python benchmark/openrouter_runner.py run --phase pilot --execute --max-requests 200 --output tmp/openrouter-pilot-001
```

A pilot schedules 27 attempts per model. To compare several explicit models, repeat `--model`:

```bash
python benchmark/openrouter_runner.py run --model vendor/model-one:free --model vendor/model-two:free --phase smoke
```

Those two IDs are placeholders; substitute current catalog entries. Each model has separate results and independent sessions. All selected models share the run's request cap. Prefer one model per run initially to simplify quota management and failure diagnosis.

## What gets recorded

- A frozen plan, sampling settings, tool schemas, system prompt, selected catalog records, binary hash, runner hashes, and source snapshot SHA.
- Request payloads without authorization headers, raw returned JSON, model/provider identifiers when supplied, and provider-reported usage per response.
- Tool calls and errors, final source, patch, final answer, termination category, and post-session behavioral grades.
- `summary.json` with every planned attempt; unavailable usage stays null.
- `review.json` with shuffled answers and rubrics, plus a separate `review-map.json`. Review packets omit explicit model and condition labels, but answer content can still reveal the condition.

Runtime data is saved under the ignored `tmp/` directory by default. It remains local until deliberately reviewed and published. Keys are redacted before saving; response bodies from HTTP errors are not logged. No private rubric, reference implementation, or expected SHA is sent to the model. The API key is used by the HTTP controller only and stripped from the MCP subprocess environment.

Regenerate the report offline with:

```bash
python benchmark/openrouter_runner.py summarize tmp/openrouter-smoke-001
```

Do this before annotating `review.json`; regeneration replaces generated report/review files. Preserve completed human reviews separately.

## Free-only policy and reproducibility limits

Requests specify one exact `:free` catalog ID, disable provider fallbacks, require parameter support, and set maximum prompt, completion, and per-request prices to zero. No server-side paid tools or plugins are requested. The runner rejects nonzero advertised pricing and stops if a response reports nonzero cost. These controls depend on OpenRouter enforcing its documented API contract; they do not replace account-side spending controls.

The returned model may be a canonical ID without `:free`; that exact canonical equivalent is accepted and recorded. A different ID terminates the run. Provider and model weights are not pinned by a catalog ID. A provider can vary between requests even with fallbacks disabled unless explicitly pinned; report this uncertainty, retain the returned provider metadata, and avoid interpreting small latency differences as a tool effect.

Free variants have availability and rate limits. The runner applies the bounded 429 policy above and never falls back to another model. The initial experiment sends only synthetic fixture code and retrieved synthetic history to OpenRouter; it gives the model no general network or filesystem tool.

## Verification

```bash
python -m unittest discover -s benchmark -p 'test_*.py' -v
```

The fake-response tests cover free-only routing, condition-specific tools, missing/paid models, secret redaction, tool execution, request failures, length truncation, model changes, cost anomalies, turn limits, retry bounds, and failed-run summaries. The rate-limit tests cover shared pacing, atomic concurrent quota reservations, server allowances, deadlines, and absence of raw keys in the ledger. Run smoke checks separately from evaluation; inspect their transcripts before freezing a pilot configuration.

## Official references

- [Free variants and paid fallback behavior](https://openrouter.ai/docs/guides/routing/model-variants/free)
- [Models API](https://openrouter.ai/docs/api/api-reference/models/list-all-models-and-their-properties)
- [Tool calling](https://openrouter.ai/docs/guides/features/tool-calling)
- [Provider routing and price caps](https://openrouter.ai/docs/guides/routing/provider-selection)
- [Account and rate limits](https://openrouter.ai/docs/api_reference/limits)
- [Response and usage format](https://openrouter.ai/docs/api_reference/overview)
