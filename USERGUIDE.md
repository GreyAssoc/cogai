# Cog — User Guide

Practical handbook for using cog day-to-day: chatting, dispatching
agents, changing models, authoring your own agents/skills/gears, and
keeping the install up to date.

> Looking for installation? See **[DEPLOY.md](../cogai/DEPLOY.md)** in
> the public distribution. This guide assumes cog is already running
> on your Telegram (and optionally Discord) bot.

---

## 1. The 30-second mental model

You talk to cog in Telegram. Every message you send goes through one
of three paths:

| You typed… | Goes to… | Visible header |
|---|---|---|
| Plain text | Default chat (your `COG_PERSONA_NAME`, e.g. "Andy") | `🤖 Andy` |
| `@<agent_name> <prompt>` | That named agent | `🤖 @<name>` |
| `/agent <name> <prompt>` | Same — explicit command form | `🤖 @<name>` |

The reply header tells you exactly which entity answered. If you
typed `@seo` and got `🤖 Andy`, the dispatch fell through — the
agent name wasn't found in your store (pre-v0.3.15 silent-loss bug,
or just a typo).

---

## 2. Discoverability — what's available right now

Four catalogue commands. Each lists what cog can actually do today
in your install (no doc-reading required):

| Command | What it shows |
|---|---|
| `/agents` | Every agent you can dispatch (built-in + your custom) |
| `/skills` | Every loaded skill (procedural knowledge auto-triggered on matching prompts) |
| `/gears` | Every gear (tool the agent loop can call), grouped by tier |
| `/chains` | Declarative workflows — **v2 feature, not yet shipped** |

For deeper inspection: `/agent show <name>` (full details on one
agent), `/provider list` (configured providers + their model
catalogues), `/help` (the full command surface).

---

## 3. Dispatching agents

Two forms — both work, pick whichever feels natural:

```
@seo audit https://example.com/pricing
@cog-coder add a debounce to the search input in app.tsx
@researcher what's the current state of EU AI Act compliance for SMBs?
```

```
/agent seo audit https://example.com/pricing
/agent cog-coder ...
```

The reply will be prefixed with the agent name so you can confirm
the dispatch routed correctly.

**Teams** (multi-agent dispatch) work the same way:
```
@code-review-team review my changes in pkg/api/
/team code-review-team review pkg/api/
```

---

## 4. Changing the model an agent uses

Three approaches depending on what you want:

### 4a. Quick per-agent change (runtime — survives restarts)

In Telegram:
```
/agent edit seo model=deepseek_v4_pro
/agent edit cog-coder provider=anthropic model=claude-opus-4-7
/agent edit code-pessimist provider=gemini model=gemini-3.5-flash
```

The change is stored in your AgentStore row and overrides whatever
the source file or built-in default said. Works on ANY agent —
built-in or custom.

**Friendly forms accepted (v0.3.16+):**
```
/agent edit seo model=deepseek v4 pro       # normalised → deepseek_v4_pro
/agent edit seo model="DeepSeek v4 Pro"     # explicit quotes also fine
/agent edit seo description=A great agent   # multi-word values work unquoted
```

To see the current model: `/agent show <name>`.
To list all available model IDs: `/provider list`.

### 4b. Permanent change in a custom agent file

For agents you wrote yourself (lives in `agents/<name>.md`):

```bash
cd <your cog install dir>
$EDITOR agents/seo.md
# change the frontmatter:
#   provider: openai
#   model: gpt-5
docker compose up -d --force-recreate gateway
```

On next first-message-per-user the seeder re-upserts from the file.

**Caveat:** if you've ever run `/agent edit seo model=…` for that
agent, your runtime edit wins — the seeder respects user edits and
doesn't overwrite them. To reset to whatever the file says, either:
- `/agent edit seo model=<the-file-value>` (manually match it), or
- `/agent delete seo` then restart the gateway (full re-seed)

### 4c. Built-in agents (compiled into the binary)

The 37 built-ins (`cog-coder`, `researcher`, `writer`, `planner`,
`code-purist`, …) aren't editable as files on your host. Use the
runtime override pattern (#4a) for those.

### Quick reference: common model IDs

Get the authoritative list from `/provider list`. Common picks:

| Provider | Model ID | Use it for |
|---|---|---|
| `anthropic` | `claude-sonnet-4-6` | Balanced default, agentic work |
| `anthropic` | `claude-opus-4-7` | Hardest reasoning, highest cost |
| `anthropic` | `claude-haiku-4-5` | Cheap fast turns |
| `openai` | `gpt-5` | OpenAI's flagship |
| `gemini` | `gemini-3.5-flash` | Cheapest grounded-search-capable |
| `gemini` | `gemini-3.1-pro` | Highest quality at moderate cost |
| `deepseek` | `deepseek_v4_pro` | Cheap + strong reasoning |
| `deepseek` | `deepseek-reasoner` | Test-time-compute reasoning |
| `xai` | `grok-4` | Fast + cheap with web search built in |
| `moonshot` | `kimi-k2` | Strong long-context summarisation |
| `qwen` | `qwen3-max` | Alibaba's flagship |
| `zai` | `glm-4.6` | Cheapest tier for noisy tool-use loops |

---

## 5. Creating a custom agent

Two paths: a guided wizard, or by hand in a file.

### 5a. Guided wizard (Recommended)

In Telegram: `/agent create` (no args). cog walks you through
name → description → provider → model → persona → gears → system
prompt, then upserts. Type `/cancel` at any step to abort.

### 5b. By hand in a `.md` file

```bash
cd <your cog install dir>
cp agents/template.md agents/my_assistant.md
$EDITOR agents/my_assistant.md
docker compose up -d --force-recreate gateway
```

A minimal `agents/my_assistant.md`:

```markdown
---
name: my_assistant
description: Short one-liner that shows up in /agent list.
provider: anthropic
model: claude-sonnet-4-6
persona: warm
gears: read, write, fetch, web_search
---

You are a focused, helpful assistant.

Lead with the agent's single most-important responsibility.
Specify working style ("cite your sources", "quantify where possible").
Set boundaries ("don't speculate beyond the data").
Specify output shape if relevant.
```

**Frontmatter fields:**

| Field | Required? | Notes |
|---|---|---|
| `name` | ✅ | snake_case, unique per user. Becomes `@my_assistant` in Telegram. |
| `description` | recommended | One line — shown in `/agent list`, used by `dispatch_agent` for handoff routing. |
| `provider` | ✅ | One of: `anthropic`, `openai`, `gemini`, `deepseek`, `xai`, `qwen`, `moonshot`, `zai`. |
| `model` | ✅ | Model id for that provider. `/provider list` for the catalogue. |
| `persona` | optional | One of 8 presets (`warm`, `stoic`, `purist`, `pragmatist`, `pessimist`, `thorough`, `neutral`, `witty`). Defaults to your global persona. |
| `gears` | optional | Comma-separated allowlist. Empty = inherit the full registry. Restricting narrows the agent's surface (more predictable). |

**Free-tier quota:** 1 custom agent above the 37 built-ins. The
installer ships `agents/template.md` (which counts as 1 if left
in place) and `agents/seo.md` (the worked example) — you'll be at
2, one above the cap. Solutions: delete or rename `template.md`,
or upgrade to Pro/Teams.

---

## 6. Creating a custom skill

**Skills** are short procedural-knowledge fragments cog auto-triggers
on matching prompts (or you can explicitly invoke via the
`skill_invoke` gear). Lighter weight than agents — no model
override, no gear allowlist, just an instruction.

### 6a. Via Telegram (Recommended)

```
/skill create
```

(There's also a meta-skill: `create_skill` walks you through it.)

### 6b. By hand in a YAML file

```bash
cd <your cog install dir>
$EDITOR skills/custom/my_skill.yaml
docker compose up -d --force-recreate gateway
```

A minimal `skills/custom/my_skill.yaml`:

```yaml
cog_skill: v1
name: my_skill
description: Triggered when the user asks about <X>; provides <Y>.
triggers:
  - "<keyword phrase>"
  - "<another phrase>"
instructions: |
  When this skill fires, do <X>.
  Specific steps:
  1. ...
  2. ...
  Always end with <Y>.
```

**Built-in skills** (shipped with cog) live in
`skills/builtin/*.yaml` and are exposed read-only so you can read
and clone them. Useful as starting points for your own — copy a
built-in to `skills/custom/`, rename, edit.

Run `/skills` in Telegram to see what's currently loaded.

---

## 7. Creating a custom gear

**Gears** are typed tools the agent loop can call: `read`, `write`,
`fetch`, `web_search`, `git`, etc. (There is no general `bash` gear —
shell access was removed from the default surface; subprocess-spawning
gears run under an operator allowlist.) Most users don't need to author
gears (the ~100 built-ins cover most needs), but if you have
an internal API or workflow worth integrating, you can ship a
declarative HTTP gear in YAML.

Authoring guide:
[**AUTHORING.md**](https://github.com/GreyAssoc/cogai/blob/main/AUTHORING.md)
in the public cogai repo (full schema, tier rules, security
constraints).

### Quick start

```bash
cd <your cog install dir>
cp gears/template.yaml gears/my_gear.yaml
$EDITOR gears/my_gear.yaml
# Edit the cog_gear: v1 declaration
docker compose up -d --force-recreate gateway
```

Validate locally before deploying:
```bash
# Download cog-gear-lint from the release for your OS:
#   https://github.com/GreyAssoc/cogai/releases/latest
cog-gear-lint gears/my_gear.yaml
```

Run `/gears` in Telegram to see what's currently registered. New
gears appear under the relevant tier heading
(`http_declared` / `webhook_declared`).

**Free-tier quota:** 3 custom gears above the built-in Tier 0
baseline.

---

## 8. Keeping cog up to date

Every cog release follows the same upgrade path.

### Standard upgrade (most users)

```bash
cd <your cog install dir>
docker compose pull                       # fetch the new image
docker compose up -d --force-recreate gateway
docker compose logs --tail=50 gateway     # verify boot
```

`--force-recreate` matters — without it, `docker compose up -d` may
keep the old container if nothing in `docker-compose.yml` changed
even though the image content did.

To check the running version: send `/help` to your bot (footer
shows the version), or:
```bash
docker compose exec gateway /cog-gateway -version
```

### Re-running the installer (when the install scaffolding changes)

When release notes mention new env vars, new bind-mounts, or new
scaffolding (skills/, gears/, agents/ added new files), re-run the
installer. It honours your existing answers and only adds what's
missing — your `.env`, `cog_mounts.yaml`, custom files, etc. all
survive:

```bash
cd <your cog install dir>
docker run --rm -it -v $PWD:/setup greyassoc/cog-installer:latest
docker compose up -d --force-recreate gateway
```

### Pinning to a specific version

By default `docker-compose.yml` references `:latest`. To pin
(recommended for production):
```yaml
# docker-compose.yml
services:
  gateway:
    image: greyassoc/cogai:vX.Y.Z    # ← change here
```

Then `docker compose pull && docker compose up -d --force-recreate`.

### Watching for new releases

- **Source:** https://github.com/GreyAssoc/cogai/releases — every tag
- **Mirror (binaries + signed SHA256SUMS):**
  https://github.com/GreyAssoc/cogai/releases — every tag
- **Docker Hub:**
  - `greyassoc/cogai:vX.Y.Z` — Telegram gateway (pin a version, not `latest`)
  - `greyassoc/cogai-discord:v0.3.X` — Discord gateway
  - `greyassoc/cog-installer:v0.3.X` — guided installer

The CHANGELOG.md in the source repo is the canonical "what
changed" surface.

---

## 9. Where things live on disk

After installation in `<cog dir>`:

```
<cog dir>/
├── .env                      # API keys + persona + paths
├── docker-compose.yml        # container layout
├── cog_mounts.yaml           # named filesystem mounts (the dirs cog can read/write)
├── agents/
│   ├── README.md             # authoring guide
│   ├── template.md           # bare-bones authoring template
│   └── *.md                  # your custom agents (each is one file)
├── skills/
│   ├── README.md             # skill authoring guide
│   ├── builtin/*.yaml        # cog's bundled skills (read-only — clone, don't edit)
│   └── custom/*.yaml         # your custom skills
└── gears/
    ├── README.md             # gear authoring guide
    ├── template.yaml         # bare-bones gear template
    └── *.yaml                # your custom gears
```

**Persistent state** (chat history, memory facts, agent overrides,
schedules, traces) lives in Postgres, owned by the `cog_pgdata`
Docker volume. Survives container restarts; lost on
`docker compose down -v`.

---

## 10. Troubleshooting

### `@seo` (or any agent) routes to default chat

Check `/agent list` — if the agent isn't there, the seed step
didn't load it. Inspect logs:
```bash
docker compose logs gateway 2>&1 | grep -iE "seed|custom_agent"
```
Common cause (fixed in v0.3.15): a parse error on one file would
silently drop the whole catalogue. Upgrade to v0.3.15+ and re-run
the gateway.

### `/agent edit seo model=foo` says "Bad field"

Re-check the field name (the error message now lists the
accepted fields). For models, run `/provider list` to see valid
IDs for each configured provider. Multi-word model names work
unquoted in v0.3.16+ — `model=DeepSeek v4 Pro` becomes
`deepseek_v4_pro`.

### Bot is silent / "🤔 cogging …" never resolves

Check connectivity to the configured provider:
```bash
docker compose logs gateway 2>&1 | grep -iE "provider|api|timeout|error" | tail -20
```
Common causes: invalid API key, rate limit, provider outage.

### A custom skill / gear / agent edit doesn't take effect

The seeder runs per-user on first message per process. Restart the
container to clear the seed bitmap:
```bash
docker compose up -d --force-recreate gateway
```
Then send any message in Telegram to re-fire the seed.

### Free-tier quota warnings

If `/agent list` shows fewer custom agents than you have on disk,
the Free-tier quota (1 custom agent) is rejecting the rest. Logs:
```bash
docker compose logs gateway 2>&1 | grep -iE "quota|tier"
```
Either prune to one custom agent or upgrade tier.

---

## 11. Reference

- **CHANGELOG.md** — what changed in each release
- **TIERS.md** — Free/Pro/Teams feature matrix
- **DEPLOY.md** — installation + env vars + deployment topologies
- **AUTHORING.md** (cogai repo) — full gear-authoring contract
- **AGENTS.md** — agent architecture deep-dive
- **SKILLS.md** — skill system architecture
- **GEARS.md** — gear model (why typed Go funcs, not MCP)

---

*Last updated: 2026-06-30 (v0.3.16 — friendlier `/agent edit`,
model name normalisation, this user guide).*
