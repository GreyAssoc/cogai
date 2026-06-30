# cog — public distribution

> Cog is a Go-native AI agent system you run yourself. This is the
> public distribution surface: README, install instructions, the
> public gear-authoring contract, and a [Releases page](https://github.com/GreyAssoc/cogai/releases)
> hosting the compiled binaries.
>
> The source lives in a private dev repo; you don't need it to run
> cog Free.

---

## What cog gives you

A production-grade AI agent system that runs on your hardware:

- **8 model providers** — Anthropic, OpenAI, Google Gemini,
  DeepSeek, xAI, Qwen, Moonshot, z.ai. BYO keys; switch
  per-message.
- **Persistent memory** — pgvector-backed facts that survive
  restarts. Cross-session, cross-project. Always free.
- **73 typed gears** — file I/O, git, build, test, shell, web
  fetch, `.docx` / `.xlsx` / `.pdf` / `.pptx` read+write, Google
  Workspace, calendar, math, image generate/describe, audio
  transcribe, text-to-speech, web search across multiple
  backends, and more.
- **33 built-in agents** — `cog-coder` (15 languages, flagship),
  plus `general` / `researcher` / `writer` / `planner`, plus 28
  single-perspective code-review specialists.
- **Forensic-grade audit trail** — every turn is a typed Postgres
  row (cost, provider, model, tokens including provider-side
  reasoning tokens, failure-class as columns). Query with
  Metabase / Looker / `psql`.
- **Local-first, BYO-keys, privacy-preserving, licence-light** —
  cog runs on your machine, you supply your own model API keys,
  your prompts + data live in your own Postgres, and the Free
  tier needs no licence file, no signup, no account.

## Built-in agents

All 37 agents below are bundled in the binary and available **free for every tier**. Invoke with `@<name>` in your bot, or set the agent in your session config. Paid tiers don't gate individual agent access — Pro+ gates *coordinated* execution (Council, orchestrator, agent teams).

### General-purpose (9)

| `@name` | Role |
|---|---|
| `general` | Default cog assistant. Friendly, honest, hands off to a specialist when one fits better. |
| `cog-coder` | Flagship coding agent. 15 languages (bash, cpp, csharp, frontend, go, java, javascript, kotlin, php, python, ruby, rust, sql, swift, typescript) via language-specialist prompts auto-loaded by file extension. |
| `researcher` | Web-search + cross-source synthesis with cited findings. |
| `writer` | Long-form prose, structured docs, editing. Audience-aware, voice-preserving. |
| `planner` | Structured execution plans — ordered steps with verification, dependencies, rollback. Drives plan mode. |
| `lookup` | Factual-lookup specialist. Wikipedia → arXiv → web_search fallback for one-shot factual questions. |
| `fact-check` | SUPPORTED / CONTRADICTED / INCONCLUSIVE on a single claim, cited from ≥2 independent sources. |
| `media-reader` | Single file → right extractor (pdf / xlsx / office / image / audio) → plain text. |
| `senior-reviewer` | Opt-in second pass that verifies / refines a primary model's draft. Higher accuracy at extra cost. |

### Code review — cross-language (10)

| `@name` | Concern | Tier |
|---|---|---|
| `security` | Vulnerability analysis, OWASP top 10, dependency risk | 1 VETO |
| `compliance` | GDPR / regulatory / lawful basis | 1 VETO |
| `data-integrity` | Type strictness, constraint coverage | 1 VETO |
| `cross-language-security` | FFI / marshalling boundary security | 1 VETO |
| `error-handling` | Error wrapping, fallback discipline, silent-failure detection | 2 |
| `logging-audit` | Trace coverage without secret leakage | 2 |
| `design-adherence` | Adherence to stated design / architecture | 2 |
| `performance` | Hot paths, allocation, query plans | 3 |
| `architect` | Module boundaries, dependency direction, layering | 3 |
| `reviewer` | General code review — catches what specialists miss | 3 |

### Code review — Go (5)

| `@name` | Concern | Tier |
|---|---|---|
| `go-purist` | Idiomatic Go; would Rob Pike approve? | 3 |
| `go-pragmatist` | Ship-vs-perfect tradeoffs in Go | 3 |
| `go-pessimist` | Race conditions, resource leaks, what can fail | 2 |
| `go-security` | Go-specific vulns, panic propagation, supply chain | 1 VETO |
| `go-hacker` | Adversarial probing of Go code | 2 |

### Code review — Frontend (6)

| `@name` | Concern | Tier |
|---|---|---|
| `fe-purist` | Semantic HTML, vanilla ES2020+ correctness | 3 |
| `fe-pragmatist` | Browser-compat vs. cleanliness | 3 |
| `fe-pessimist` | Async edge cases, event-handler leaks, DOM resilience | 2 |
| `fe-security` | XSS, CSRF, CSP, supply chain | 1 VETO |
| `fe-hacker` | Browser-side adversarial probing | 2 |
| `fe-functional` | Functional correctness, event flow, state mutation | 3 |

### Code review — Mobile / Kotlin (4)

| `@name` | Concern | Tier |
|---|---|---|
| `kotlin-purist` | Idiomatic Kotlin, structured concurrency | 3 |
| `mobile-security` | Android security model, permission abuse, supply chain | 1 VETO |
| `mobile-hacker` | Adversarial probing of mobile code | 2 |
| `android-mobile-coding` | Jetpack Compose, Hilt, Room patterns | 3 |

### Code review — Integration (3)

| `@name` | Concern | Tier |
|---|---|---|
| `api-guardian` | API contract stability, version compatibility | 2 |
| `data-flow` | End-to-end data flow correctness | 2 |
| `failure-modes` | Failure injection, recovery paths | 2 |

### Research helper (1)

| `@name` | Concern | Tier |
|---|---|---|
| `ui-ux-researcher` | UI/UX patterns, accessibility | 3 |

### The coder-review-coder loop

The canonical Free-tier coding workflow. Reviewers don't have to be ganged up into a team to be useful — a single review pass between coder rounds catches most issues.

```
@cog-coder       writes / fixes code
@<review-agent>  critiques (any single agent, e.g. @go-security)
@cog-coder       revises based on the critique
@<other-agent>   (optional) second-pass review
```

This loop ships as a CI-tested regression suite. See [TIERS.md §7.3](https://github.com/GreyAssoc/cog/blob/main/TIERS.md) on the source repo for the full agent reference.

## Quick install — guided setup via Docker (recommended)

If you have Docker installed, this is the fastest path. One
interactive container walks you through the same setup the native
installer does — Telegram bot token, allowed user IDs, your model
provider API key(s), default model, install location — and writes
`docker-compose.yml` + `.env` + `cog_mounts.yaml` to your current
directory:

```bash
mkdir cog && cd cog
docker run --rm -it -v $(pwd):/setup greyassoc/cog-installer:v0.3.7
docker compose up -d
```

Open Telegram, find your bot, send `/help`. The first run pulls
~120 MB (gateway image + bundled `pgvector` Postgres) and
typically completes in under 60 seconds on a fast connection.

The Discord variant of the gateway is a separate image and gets
wired in automatically by the installer if you supply a Discord
bot token:

```bash
docker pull greyassoc/cogai-discord:v0.3.7
```

All three images publish multi-arch (`linux/amd64` + `linux/arm64`).

### Hand-rolled compose (advanced)

If you want to skip the installer and write your own
`docker-compose.yml`, pull the gateway image and bundle it with
`pgvector/pgvector:pg16`. See [`DEPLOY.md`](https://github.com/GreyAssoc/cogai/blob/main/DEPLOY.md)
for the required env vars (Telegram token, allowed user IDs, at
least one model provider key, Postgres URL).

```bash
docker pull greyassoc/cogai:v0.3.7
```

## Install via the native installer (no Docker prerequisite)

If you don't already have Docker installed and prefer a fully
guided setup (the native installer prompts you through Docker
installation too where needed), grab the platform binary from
[Releases](https://github.com/GreyAssoc/cogai/releases/latest):

| Platform | File |
|---|---|
| Windows (Intel / AMD) | `cogai-installer-windows-amd64.zip` |
| macOS (Intel) | `cogai-installer-darwin-amd64.tar.gz` |
| macOS (Apple Silicon) | `cogai-installer-darwin-arm64.tar.gz` |
| Linux (Intel / AMD) | `cogai-installer-linux-amd64.tar.gz` |
| Linux (ARM, e.g. Raspberry Pi 4) | `cogai-installer-linux-arm64.tar.gz` |

Then:

1. Extract; double-click the launcher (`Install Cog.command`,
   `install-cog.sh`, or `install-cog.bat`).
2. Have to hand: a Telegram bot token (and/or a Discord bot
   token), an API key for at least one model provider, your
   email.
3. When the installer reports `✓ cog is running.`, open
   Telegram / Discord and message your bot `/help`.

For a guided walk-through covering the Telegram and Discord bot
setup steps, see [`DEPLOY.md`](https://github.com/GreyAssoc/cogai/blob/main/DEPLOY.md).

### Verifying the download

Every release ships a `SHA256SUMS` file alongside the binaries.
Verify your download integrity with:

```bash
sha256sum -c SHA256SUMS --ignore-missing
```


## Built-in skills (11, all tiers)

Skills are procedural-knowledge YAMLs auto-injected into the system prompt when a regex trigger matches the user's prompt. They sit between gears (one callable) and agents (long-lived identity): a short recipe activated only when the model is about to do the matching kind of work. **All 11 ship in the binary and are free for every tier.** Add up to 3 custom skills on Free, unlimited on Pro+.

Auto-trigger is on by default; toggle off globally via `COG_SKILLS_AUTO=false`. The model can also load any skill explicitly via the `skill_invoke(name)` gear.

| Skill | Fires on prompts like… | What it does |
|---|---|---|
| `git_commit_workflow` | "commit these changes", "let's commit", "git commit" | Commit safely — status, diff, scoped add, why-not-what message. |
| `code_review_protocol` | "review this PR", "check this code", "feedback on this" | Layered code review — correctness, security, tests, edge cases, readability. |
| `research_with_citations` | "research", "look up", "find out about", "what's the latest on" | Multi-source research with explicit citations + cross-check. |
| `safe_overwrite_check` | "overwrite", "replace this file", "rewrite this config" | Read-before-write pattern — never blind-overwrite an existing file. |
| `debug_systematic` | "this is broken", "find the bug", "why does it crash" | Reproduce → isolate → test → fix → verify. No symptom-chasing. |
| `incident_response` | "production is down", "outage", "incident" | Note timing, check logs, communicate, snapshot, postmortem. |
| `check_email` | "check my email", "any new messages", "what's in my inbox" | Triage inbox into needs-attention / read-later / ignore. Offer to draft replies. |
| `check_calendar` | "check my calendar", "what's on today", "do I have meetings" | Today's schedule with conflict / back-to-back / VIP flags + meeting-prep offer. |
| `compose_email_carefully` | "send an email", "draft a reply", "compose" | Pre-send checklist — subject, recipients, tone, attachments, no auto-send. |
| `meeting_prep` | "prepare for my meeting", "brief me on my next call" | Brief for an upcoming meeting from calendar + email + memory. |
| `create_skill` | "create a new skill", "how do I author a skill" | Meta-skill — walks you through name, triggers, procedure, required gears, save path, and a smoke test for a new custom skill. |

### Storage layers (highest precedence first)

```
<workdir>/.cog/skills/*.yaml   — per-project overrides (refused if inside a FileWriteRoot)
~/.cog/skills/*.yaml           — operator-authored (the recommended layer)
engine/skills/builtin/*.yaml   — the 11 built-ins, embedded into the binary
```

Same-name skills in a higher layer fully replace lower layers — operators override a built-in by dropping `~/.cog/skills/git_commit_workflow.yaml` rather than patching the binary. Name normalisation (TrimSpace + ToLower) means a near-alias like "Create_Skill " still resolves to the built-in `create_skill` key. The `<workdir>/.cog/skills/` layer is refused in default deployments because the agent's `write` gear can reach it — keep custom skills in `~/.cog/skills/` on the host.

## Built-in gears (89, all tiers)

Gears are typed Go functions the agent dispatches as tools. **None of the 89 gears below are tier-gated** — every gear in the binary is available on every tier including Free. What Pro+ unlocks is the *quota for adding your own* (custom_gears_max — Free 3, Pro unlimited) and the orchestration gears that coordinate multiple agents.

### File I/O & search

| Gear | Purpose |
|---|---|
| `read` | Read a file (text, code, or extracted `.docx` / `.xlsx` / `.pdf` / `.pptx`). Smart-truncates large files. |
| `write` | Write or overwrite a file. Path-confined to `FileWriteRoots`. |
| `edit` | Targeted edit — replace a string, patch a line range. Cheaper than full write. |
| `list` | List directory contents with type, size, mtime. |
| `find_files` | Glob across the read roots. Honours `.cog-ignore`. |
| `search` | Regex content search across read roots. |
| `file_info` | Stat a single path — size, mtime, type, content sample. |
| `path_resolve` | Resolve a "label"-style path (`@active-jobs/foo.docx`) to a real fs path. |
| `extract` | Generic text extractor; routes by extension to the right reader. |
| `diff` | Compute a structured diff between two files or strings. |
| `mkdir` / `move` / `copy` / `remove` / `touch` | Filesystem ops, write-root confined, pending-op confirmation for destructive actions. |

### Shell, build, VCS

| Gear | Purpose |
|---|---|
| `bash` | Execute a shell command. Allowlist-required in production; `COG_DISABLE_BASH=true` to disable. |
| `run` | Project's test command (`project.commands.test`). Captures exit + duration + output. |
| `git` | Structured `git status` / `git diff` / common subcommands. No shell pipe risk. |
| `code_build` | Project's build command. Returns exit, duration, captured excerpt. |
| `code_test` | Project's test suite. |
| `code_status` | Aggregate of git status + last build/test status. |
| `code_diff` | Pending diff vs HEAD or a named ref. |
| `code_lint` | Run the configured linter. Returns structured findings when parseable. |
| `code_format` | Run the configured formatter on a file. |

### Code intelligence (cog's static analysis index)

| Gear | Purpose |
|---|---|
| `code_outline` | File's top-level symbol map — funcs, types, consts with line ranges. |
| `code_definition` | Where a symbol is defined — file + line range + doc excerpt. |
| `code_callers` | Direct callers of a symbol. Graph-derived; no grep false-positives. |
| `code_callees` | What a symbol calls. |
| `code_imports` | Outgoing imports of a file. |
| `code_impact` | Transitive callers of a symbol up to depth N — full blast radius of a change. |
| `code_path` | Find a reference path between two symbols. |
| `code_callers_global` / `code_impact_global` / `code_path_global` | Same three queries across the cross-project index. Operator opt-in. |
| `code_exec` | Execute a snippet (Python or JavaScript) via the provider's native code execution where available. 30s cap locally. |

### Web

| Gear | Purpose |
|---|---|
| `fetch` | HTTP(S) GET with SSRF guard (blocks private / loopback / link-local). |
| `web_search` | Multi-backend web search (Brave / Bing / DDG / Gemini-grounded) with key-aware backend ordering. |
| `wikipedia` | Direct Wikipedia search + extract — prefer over `web_search` for encyclopedic questions. |
| `arxiv` | arXiv paper search + metadata. |
| `weather` | Current + forecast weather via Open-Meteo. |
| `maps` | Geocode + directions via OSM-backed services. |
| `youtube_transcript` | Pull a YouTube video's transcript by URL. |
| `fx_rate` | Live FX rate between two currencies. |
| `stock_quote` | Live stock quote by ticker. |

### Media + document extraction & generation

| Gear | Purpose |
|---|---|
| `pdf_extract` | Extract text + structure from a PDF. |
| `xlsx_extract` | Extract structured cells + ranges from `.xlsx` / `.xlsm`. |
| `audio_transcribe` | Transcribe an audio file (mp3 / m4a / wav / etc.) via Gemini. |
| `image_describe` | Caption / describe an image. |
| `image_generate` | Generate an image from a prompt. |
| `text_to_speech` | Synthesise speech to an audio file. |
| `docgen` | Generic document generation. |
| `office_write` | Create `.docx` / `.xlsx` / `.pptx` / `.pdf` from structured content. |

### Google Workspace (BYO OAuth)

| Gear | Purpose |
|---|---|
| `gmail_list_unread` / `gmail_search` / `gmail_get` / `gmail_thread` | Read inbox: list unread, search, fetch a message, fetch a thread. |
| `gmail_mark` / `gmail_send` | Mark read/unread/important; send a message. |
| `calendar_list_events` / `calendar_get` / `calendar_find_free` | Read calendars: list, fetch, find free slots across calendars. |
| `calendar_create` / `calendar_update` / `calendar_delete` | Write calendar: create / update / delete with pending-op confirmation. |
| `drive_search` / `drive_get_metadata` / `drive_read` | Read Drive: search, metadata, content-as-text. |
| `drive_upload` / `drive_share` | Write Drive: upload a file, share with another user. |
| `sheets_read` / `sheets_list` | Read Sheets: cells/ranges, list sheets in a workbook. |
| `sheets_write` / `sheets_append` / `sheets_clear` | Write Sheets: replace, append, clear ranges. |

### Memory + state

| Gear | Purpose |
|---|---|
| `remember` | Save a persona fact to pgvector-backed memory (cross-session, cross-project). |
| `forget` | Remove a previously-saved fact. |

### Cron + orchestration

| Gear | Purpose | Tier |
|---|---|---|
| `cron_schedule` | List, add, remove, enable / disable, or run cron jobs. | Free |
| `dispatch_agent` | Hand off the turn to a specialist built-in agent (e.g. `@researcher`). | Free |
| `subagent` | Run a sub-cog with a separate context window for a focused subtask. | **Pro+** |
| `task_start` / `task_status` / `task_result` / `task_update` / `task_list` / `task_cancel` | Async task management — fire-and-forget long-running gears that survive turn boundaries. | Free |

### Specialised (UK party-wall surveying)

| Gear | Purpose |
|---|---|
| `grey_clients_list` | List active client folders under the surveyor workspace. |
| `grey_clients_find` | Fuzzy-find a client folder by name. |
| `grey_docs_generate` | Generate surveyor documents (notices, letters, reports) into a timestamped subdir. |
| `grey_docs_serve_notices` | Multi-recipient notice service flow. |

### Time

| Gear | Purpose |
|---|---|
| `date_time` | Current time, timezone, date arithmetic — the agent should never guess "today's date". |

---

## Tier matrix — what each tier unlocks

The principle, from `TIERS.md §1`: **never gate things that make cog useful as a coding agent.** Memory, plan mode, cron, code intelligence, every model provider, every built-in agent, every built-in gear are always Free. Paid tiers gate *coordinated execution* and *unlimited extensibility* — features that compound with scale, not features needed to ship one project.

| Capability | **Free** *(shipping)* | **Pro** *(planned)* | **Teams** *(planned)* |
|---|---|---|---|
| **Seats** | 1 | 1 | unlimited |
| **All 8 model providers** | ✓ | ✓ | ✓ |
| **All models, no token markup** | ✓ | ✓ | ✓ |
| **All 89 built-in gears** | ✓ | ✓ | ✓ |
| **All 37 built-in agents** | ✓ | ✓ | ✓ |
| **All 11 built-in skills** | ✓ | ✓ | ✓ |
| **Persistent pgvector memory** | ✓ | ✓ | ✓ |
| **Plan mode + stuck-detector** | ✓ | ✓ | ✓ |
| **Cron scheduler** | ✓ | ✓ | ✓ |
| **Cancellation (`/stop`)** | ✓ | ✓ | ✓ |
| **Streaming progress** | ✓ | ✓ | ✓ |
| **Auto-compaction** | ✓ | ✓ | ✓ |
| **Persona overlays (8 × M/F/N)** | ✓ | ✓ | ✓ |
| **Coder-review-coder loop** | ✓ | ✓ | ✓ |
| **Custom gears beyond Tier 0** | 3 | **unlimited** | **unlimited** |
| **Custom agents beyond built-ins** | 1 | **unlimited** | **unlimited** |
| **Custom skills beyond built-ins** | 3 | **unlimited** | **unlimited** |
| **Council (4-agent parallel + chair + refinement)** | ✗ | **✓** | **✓** |
| **Orchestrator (`subagent` gear)** | ✗ | **✓** | **✓** |
| **Bundled agent teams** (go-dev-team, fe-sec-team, …) | ✗ | **✓** | **✓** |
| **Self-imposed spend cap UI** | ✗ | **✓** | **✓** |
| **Audit / governance dashboards** | ✗ | self-only | **full multi-user** |
| **Per-user policy mutation (budget caps, allowed providers, quotas)** | ✗ | self-only | **admin over users** |
| **OIDC SSO** | ✗ | ✗ | **✓** |
| **GDPR Art. 15 subject-access endpoint** | ✗ | ✗ | **✓** |
| **Reporting endpoints (usage / violations / cost outliers / transparency)** | ✗ | ✗ | **✓** |
| **Trace retention** | 30 days | 365 days | configurable |
| **Channels** | Telegram + Discord | + WhatsApp + web chat + HTTP API + IDE/CLI | + on-prem connector |
| **Source-access rider (audit the engine under NDA)** | ✗ | ✗ | **✓** |
| **No phone-home** | ✓ | daily check-in (licence_id only) | configurable: online or fully offline |
| **Licence file required?** | no | yes (Ed25519-signed) | yes (Ed25519-signed, org-bound) |

### Free is a credible standalone product

Free isn't a teaser. You get a real coding agent that beats most paid alternatives on the coder-review-coder loop, **all** model providers (BYO keys, no token markup), persistent memory across restarts, cron, plan mode, 37 built-in agents, 89 typed gears, and 30-day forensic-grade audit retention. Everything you need to ship one project end-to-end.

### Pro unlocks parallel coordination + unlimited extensibility

The features that compound with scale: running 4 reviewers in parallel via Council, dispatching a sub-cog for a focused subtask, bundling reviewers into pre-configured teams, building unlimited custom gears for your own SaaS surface, hosting on WhatsApp / web / HTTP / your IDE.

### Teams adds the governance plane

For multi-user deployments: OIDC, per-user budget caps enforced as integers (no float drift), GDPR subject-access, audit dashboards across the whole org, on-prem connector, source-access rider for procurement due diligence.

Pricing for Pro / Teams is deliberately deferred — see [getcog.ai/pricing](https://getcog.ai/pricing). The Free tier is perpetual, single-seat, BYO-keys; no signup or licence file required.

## Authoring declarative gears

Cog runs your custom HTTP-backed gears as **Tier 1 declarative
gears** — YAML or JSON declarations that the engine loads at
runtime and dispatches with the same permission gates as compiled
gears.

- [`cog_gear.v1.schema.yaml`](./cog_gear.v1.schema.yaml) — canonical
  meta-schema (YAML for readability).
- [`cog_gear.v1.schema.json`](./cog_gear.v1.schema.json) — JSON
  Schema for editor integration (VS Code YAML extension, IntelliJ).
- [`AUTHORING.md`](./AUTHORING.md) — Tier 1 (HTTP) and Tier 2
  (webhook) authoring guide.
- [`examples/`](./examples/) — reference gear declarations: Notion,
  Stripe, GitHub, Linear.

To validate a declaration before submitting it, download the
`cog-gear-lint` binary for your platform from
[Releases](https://github.com/GreyAssoc/cogai/releases) and run it
against your file:

```bash
cog-gear-lint my_gear.yaml
# OK: my_gear.yaml
```

## Licence

Closed-source proprietary. See [LICENSE](./LICENSE).

The Free tier is perpetual, single-seat, BYO-keys; no licence file
or signup required. Paid tiers (Pro / Teams) require an
Ed25519-signed licence file issued on subscription.

The schema files (`cog_gear.v1.schema.{yaml,json}`) and the example
gear declarations are published as the public contract for
Tier 1 / Tier 2 gear authoring. They may be referenced and used by
third-party editor extensions, validators, and authoring tools
under the same terms as the rest of cog.

---

**Issues & support:** [github.com/GreyAssoc/cogai/issues](https://github.com/GreyAssoc/cogai/issues) · support@getcog.ai
**Domain:** [getcog.ai](https://getcog.ai)
**Contact:** Steve Whitehead — steve.w@greyandassociates.co.uk

