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

## Quick install — Docker (recommended)

If you have Docker, the fastest path is to pull the published
image and start a single-container deployment:

```bash
docker pull greyassoc/cogai:v0.3.3
```

For a full deployment (gateway + bundled Postgres), use the
docker-compose template from the installer (see below) — it wires
up the persistent volume + healthcheck + restart policy for you.

The Discord variant is a separate image:

```bash
docker pull greyassoc/cogai-discord:v0.3.3
```

Multi-arch: both images publish for `linux/amd64` + `linux/arm64`.

## Install via the installer (~5 minutes, all platforms)

If you don't already have Docker installed and prefer a guided
setup, grab the installer for your platform from
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

## Tiers

| Tier | Status | Pitch |
|---|---|---|
| **Free** | **shipping** | 1 seat, real coding agent, 3 custom gears, 1 custom agent, Telegram + Discord, 30-day trace retention. **Persistent memory always free.** All 8 model providers, all models, no token markup. |
| Pro | planned | Parallel agent coordination (Council + orchestrator + agent teams), unlimited custom gears/agents, private professional interfaces (WhatsApp + web chat + HTTP API + IDE/CLI host), self-imposed spend cap, 365-day retention. |
| Teams | planned | Multi-user admin plane, OIDC SSO, per-user policy mutation, on-prem connector. Governance-priced, not chat-priced. |

Full tier-vs-feature matrix at [getcog.ai/pricing](https://getcog.ai/pricing).
Pricing for paid tiers is deliberately deferred until Free has
shipped and we have real signal.

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

