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

- **7 model providers** — Anthropic, OpenAI, Google Gemini, DeepSeek,
  xAI, Qwen, Moonshot. BYO keys; switch per-message.
- **Persistent memory** — pgvector-backed facts that survive
  restarts. Cross-session, cross-project. Always free.
- **Typed tools (gears)** — file I/O, git, build, test, shell, web
  fetch, `.docx` / `.xlsx` / `.pdf` / `.pptx` read+write, Google
  Workspace, calendar, math. ~30 built-in.
- **33 built-in agents** — `cog-coder` (15 languages, flagship),
  plus `general` / `researcher` / `writer` / `planner`, plus 28
  single-perspective code-review specialists.
- **Telegram + Discord** — both channels ship in Free.
- **Audit-friendly** — every turn is a typed Postgres row (cost,
  provider, model, tokens, failure-class as columns). Query with
  Metabase / Looker / `psql`.

## Install (~5 minutes)

1. Grab the installer for your platform from
   [Releases](https://github.com/GreyAssoc/cogai/releases):

   | Platform | File |
   |---|---|
   | Windows (Intel / AMD) | `cog-installer-windows-amd64.zip` |
   | macOS (Intel) | `cog-installer-darwin-amd64.tar.gz` |
   | macOS (Apple Silicon) | `cog-installer-darwin-arm64.tar.gz` |
   | Linux (Intel / AMD) | `cog-installer-linux-amd64.tar.gz` |
   | Linux (ARM, e.g. Raspberry Pi 4) | `cog-installer-linux-arm64.tar.gz` |

2. Extract; double-click the launcher (`Install Cog.command`,
   `install-cog.sh`, or `install-cog.bat`).
3. Have to hand: a Telegram bot token (and/or a Discord bot token),
   an API key for at least one model provider, your email.
4. When the installer reports `✓ cog is running.`, open Telegram /
   Discord and message your bot `/help`.

For a guided walk-through covering the Telegram and Discord bot
setup steps, see [`DEPLOY.md`](https://github.com/GreyAssoc/cogai/blob/main/DEPLOY.md).

## Tiers

| Tier | Status | Pitch |
|---|---|---|
| **Free** | **shipping** | 1 seat, real coding agent, 3 custom gears, 3 custom agents, Telegram + Discord, 30-day trace retention. **Persistent memory always free.** |
| Pro | planned | Council, orchestrator, agent teams, unlimited custom gears/agents, web chat + HTTP API + IDE/CLI host, WhatsApp, 365-day retention. |
| Family | planned | 4 seats with parent-controlled spend caps + audit. |
| Teams | planned | Multi-user admin plane, OIDC SSO, per-user policy mutation, on-prem connector. |

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

## License

Closed-source proprietary. See [LICENSE](./LICENSE).

The Free tier is perpetual, single-seat, BYO-keys; no licence file
or signup required. Paid tiers (Pro / Family / Teams) require an
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
