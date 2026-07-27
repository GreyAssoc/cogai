# Cog — Manual Deployment Walkthrough

> The one-click installer at [`install/`](./install/) is the fast
> path. This document is the long path — what every prompt is doing
> under the hood, plus the bot-registration steps for Telegram and
> Discord (which the installer can't do for you).
>
> If you just want cog running and don't need to peek under the
> hood: download the installer for your OS from
> [GitHub Releases](https://github.com/GreyAssoc/cogai/releases) and
> skip to [§3](#3-discord-bot-setup) and [§2](#2-telegram-bot-setup)
> only when the installer prompts you for the bot tokens.

---

## 1. Prerequisites

| Requirement | Why |
|---|---|
| Docker Engine 24+ | Runs the gateway + bundled Postgres |
| Docker Compose v2 (`docker compose` not `docker-compose`) | Orchestrates services |
| 4 GB RAM free | Postgres + gateway + headroom |
| 10 GB disk free | Postgres data volume + container images |
| Outbound HTTPS to your model provider(s), `api.telegram.org`, `discord.com` | Provider, Telegram long-poll, Discord gateway |

No host-installed Go, Postgres, or Node. The Docker image bundles
everything.

---

## 2. Telegram bot setup

You need three things from Telegram before cog can talk to you on
that channel: a **bot token**, your **numeric user ID**, and a chat
opened with the bot.

### 2.1 Create the bot

1. Open Telegram and message [@BotFather](https://t.me/BotFather).
2. Send `/newbot`.
3. BotFather asks for a display name, then a username ending in
   `bot`. Pick anything — `MyCogBot` / `acme_assistant_bot` are
   fine.
4. BotFather replies with a token that looks like
   `7234567890:AAH...` — **save this**. It's your
   `COG_GATEWAY_TELEGRAM_TOKEN`.

### 2.2 Find your numeric user ID

Cog gates incoming messages by Telegram user ID — only IDs you list
in `COG_GATEWAY_ALLOWED_USER_IDS` can talk to your bot. Everyone
else gets silently dropped.

1. Open Telegram, message [@userinfobot](https://t.me/userinfobot).
2. Send any message; the bot replies with your numeric ID
   (e.g. `1529129983`).
3. That's the value for `COG_GATEWAY_ALLOWED_USER_IDS`. Multiple
   IDs comma-separated.

### 2.3 Start the chat

Open your own bot (the one BotFather just made) and click **Start**.
Telegram won't deliver messages until the user starts the chat once
— this is a one-shot bootstrapping step.

---

## 3. Discord bot setup

The Discord gateway is optional in Free, but it's the second
no-per-message-cost channel and adds organic reach (one user adds
your bot to their server, 50–500 other members see cog).

### 3.1 Create the Discord application

1. Open [discord.com/developers/applications](https://discord.com/developers/applications).
2. Sign in. Click **New Application**, name it (`MyCog` etc.), agree
   to the Discord Developer ToS.
3. The application page loads. Open the **Bot** tab on the left.
4. Click **Add Bot** → **Yes, do it!**.
5. Under **Token**, click **Reset Token** → **Yes, do it!**. Discord
   shows the token **once** — copy it immediately. That's your
   `DISCORD_BOT_TOKEN`. (If you miss it, you can reset again; old
   token gets revoked.)
6. Under **Privileged Gateway Intents**, enable:
   - **Message Content Intent** (cog needs to read message text).
   - (Optional) **Server Members Intent** — only if you'll use
     custom membership-aware gears later.

### 3.2 Invite the bot to a server

1. Open the **OAuth2 → URL Generator** tab.
2. Under **Scopes**, tick `bot` and `applications.commands`.
3. Under **Bot Permissions**, tick at minimum:
   - `Send Messages`
   - `Read Message History`
   - `Use Slash Commands`
   - (Optional) `Attach Files` — for artifact attachments
4. Copy the generated URL at the bottom. Paste it in a new browser
   tab. Pick the server (Guild) to invite the bot to. You need
   **Manage Server** permission on that server.
5. Discord redirects you back to the application page when done.

### 3.3 Get the Guild ID (optional)

If you want to restrict the bot to specific servers (set
`DISCORD_GUILD_ALLOWLIST`), you need each server's numeric Guild
ID.

1. In the Discord desktop app, open **User Settings → Advanced →
   Developer Mode** and toggle it on.
2. Right-click any server icon in your sidebar → **Copy Server ID**.
3. Paste that into the installer's allowlist prompt
   (comma-separated for multiple servers). Empty = every server
   the bot is invited to.

### 3.4 Start a chat

Mention the bot in any channel it has access to, or DM it directly.
DMs don't need a `@mention`; channel messages do (cog only responds
when explicitly mentioned in a server channel, to avoid wading into
every message).

---

## 4. Choose a model provider

You need an API key from at least one of:

| Provider | Get a key | Env var |
|---|---|---|
| Anthropic Claude | [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys) | `ANTHROPIC_API_KEY` |
| OpenAI | [platform.openai.com/api-keys](https://platform.openai.com/api-keys) | `OPENAI_API_KEY` |
| Google Gemini | [aistudio.google.com/apikey](https://aistudio.google.com/apikey) | `GEMINI_API_KEY` |
| DeepSeek | [platform.deepseek.com/api_keys](https://platform.deepseek.com/api_keys) | `DEEPSEEK_API_KEY` |
| xAI Grok | [console.x.ai](https://console.x.ai) | `XAI_API_KEY` |
| Qwen | [dashscope.console.aliyun.com](https://dashscope.console.aliyun.com) | `QWEN_API_KEY` |
| Moonshot | [platform.moonshot.cn/console/api-keys](https://platform.moonshot.cn/console/api-keys) | `MOONSHOT_API_KEY` |

You can supply more than one key. When more than one is configured,
set `COG_DEFAULT_PROVIDER=<name>` to pick the default; users switch
per-chat with `/provider use <name>`.

### 4.1 The optional `/clients/*` integration

Cog can mount a tree of read-only client folders (e.g. legal docs,
PDFs, project archives) at `/clients/current/` and `/clients/past/`
inside the container, with a single overlay writable folder for
the agent's scratch.

If you don't have a client tree, skip this when the installer
asks. The default install gives the agent read+write over a
`workspace/` directory in your install location, which is plenty
for most use.

---

## 5. Lay down the deployment

The installer does this for you — but if you're doing it by hand:

**cog's source is private, so there is no source tree to clone.** The
supported way to get a deployment on disk is to run the installer once —
it writes a complete `docker-compose.yml` + `.env` + `cog_mounts.yaml`
into a directory you choose, and you edit those afterwards like any
other compose stack:

```bash
# 1. Pick an install directory.
mkdir -p ~/cog && cd ~/cog

# 2. Let the installer materialise the stack here.
docker run --rm -it -v $(pwd):/setup greyassoc/cog-installer:v0.4.0

# 3. Edit .env. At minimum:
#    COG_GATEWAY_TELEGRAM_TOKEN, COG_GATEWAY_ALLOWED_USER_IDS,
#    one provider key, COG_GATEWAY_USER_EMAIL,
#    COG_HOST_WORKSPACE=/absolute/path/to/cog/workspace

# 4. Edit docker-compose.yml if needed — drop the Discord service if
#    you're Telegram-only, add client mounts if applicable.

# 5. Run.
mkdir -p workspace
docker compose up -d

# 6. Watch the gateway boot.
docker compose logs -f gateway
```

From here on the stack is ordinary Docker Compose; nothing about it
depends on cog's source. If you want to author the compose file from
scratch instead, the images and their environment contract are
documented in §7 and on
[Docker Hub](https://hub.docker.com/r/greyassoc/cogai).

When you see `gateway listening` / `polling Telegram` / similar
without an error, message `/help` to your bot.

---

## 6. Operational commands

```bash
# Stop everything.
docker compose down

# Stop + drop the Postgres volume (clean slate).
docker compose down -v

# Restart.
docker compose up -d

# Tail logs.
docker compose logs -f gateway          # Telegram
docker compose logs -f gateway-discord  # Discord (if enabled)
docker compose logs -f postgres         # database

# Rebuild after a source update.
docker compose build && docker compose up -d --force-recreate
```

The installer prints these on its way out; this section duplicates
them for the long-path users.

---

## 7. Verifying it works

In Telegram, message your bot:

```
/help
```

Expected: the command catalogue arrives within ~2 seconds.

```
what files are in /workspace?
```

Expected: a `list` gear dispatches, the bot replies with a listing
of your workspace mount.

```
/cron list
```

Expected: empty list, since no jobs are scheduled.

```
/remember my favourite colour is blue
```

Expected: `📌 Remembered: my favourite colour is blue`. Now ask
`what's my favourite colour?` — the bot recalls from the facts
store.

For Discord, **mention the bot in a server channel** (`@MyCog hello`)
or **DM it directly**. The same commands work; replies arrive in
the same channel where you triggered them.

---

## 8. Verifying audit posture

Cog writes every turn to the bundled Postgres as a typed trace row.
You can query it directly:

```bash
docker compose exec postgres psql -U cog -d cog -c   "SELECT kind, count(*) FROM cog_traces_v1 GROUP BY kind ORDER BY count DESC;"
```

Expected output after a handful of messages:

```
       kind        | count
-------------------+-------
 model_response    |    12
 gear_call         |     8
 gear_result       |     8
 session_start     |     3
 session_end       |     3
 ...
```

To see per-user spend so far:

```bash
docker compose exec postgres psql -U cog -d cog -c "SELECT * FROM cog_spend_by_user_day LIMIT 10;"
```

These columns are typed (not free-text JSON), so you can wire
Metabase / Looker / Tableau against them directly.

---

## 9. Troubleshooting

### `Forbidden: bot was blocked by the user` (Telegram)

You blocked your own bot during testing. Open the bot's chat in
Telegram, click **Unblock**.

### `401 Unauthorized` from Discord on boot

Token is wrong or revoked. Go back to
[discord.com/developers/applications](https://discord.com/developers/applications)
→ your app → **Bot** → **Reset Token**, copy the new one, replace
`DISCORD_BOT_TOKEN` in `.env`, restart: `docker compose up -d --force-recreate gateway-discord`.

### Bot doesn't respond in a Discord server channel

You probably didn't `@mention` it. Cog only responds when explicitly
addressed in server channels (DMs don't need the mention). Or the
guild isn't in `DISCORD_GUILD_ALLOWLIST` (if you set one).

### `pq: database "cog" does not exist`

Postgres init didn't run. Try `docker compose down -v && docker
compose up -d`. The `-v` drops the volume so Postgres re-initialises
with the right user/db/password.

### Gateway boots then exits with `config: missing required env vars`

The named env vars aren't in `.env`. Open `.env` and confirm the
listed ones have values. Restart: `docker compose up -d
--force-recreate`.

### Provider returns `401` / `403` / `quota exceeded`

That's a provider-side issue, not a cog issue. The provider's
dashboard is the authoritative source. Cog's `cost_usd` column is
a best-effort estimate; the provider's bill is the truth.

### Long messages get cut off in Telegram

Telegram has a 4096-char limit per message. Cog chunks longer
replies into multiple messages automatically. If the cut-off looks
mid-sentence, file an issue with the trace ID.

### Help is `?` not `/help`?

Both work. Cog accepts `?`, `.`, `help`, `/help` — all four send
the same command catalogue. The `/` autocomplete menu in Telegram
lists the canonical forms.

---

## 10. Upgrading

```bash
cd /path/to/cog/source
git pull
cd /path/to/cog
docker compose up -d --build --force-recreate
```

The bundled Postgres data volume survives the rebuild; your facts,
cron jobs, sessions, and trace history persist.

For the one-click installer flow, just re-run the installer from
your install directory — existing answers are kept; new prompts
(e.g. for the Discord setup added in this release) are appended.

---

## 11. Production posture

The defaults above are tuned for a single-operator pilot. Before
exposing the bot to other users (i.e. before Teams), walk
the production checklist in [`docs/runbook.md`](./docs/runbook.md):

- `COG_BASH_REQUIRE_ALLOWLIST=true` — despite the name, this no longer
  gates a `bash` gear (there isn't one). It constrains which *programs*
  the typed subprocess gears may invoke — `git`, the language toolchains,
  the sandboxer — matched by base name against `COG_BASH_ALLOWLIST`.
- `COG_GATEWAY_FORMAT=html`
- `PolicyFailClosed=true`
- `COG_MASTER_FORCE_SECURE_COOKIES=true` (when running cog-master)
- Behind a TLS-terminating proxy
- Audit role split (admin emails in `COG_MASTER_ADMIN_EMAILS`)
- Egress proxy if you're worried about prompt-injection exfil via
  the `fetch` gear

Free-tier single-operator deployments don't need most of this.

---

## 12. Where to go next

| You want… | Read |
|---|---|
| The tier matrix (what Free includes vs. paid) | [TIERS.md](./TIERS.md) |
| The tool model (why no MCP) | [GEARS.md](./GEARS.md) |
| The built-in agent inventory | [AGENTS.md](./AGENTS.md) |
| Production runbook | [docs/runbook.md](./docs/runbook.md) |
| To author a Tier 1 / Tier 2 declarative gear | [github.com/GreyAssoc/cogai](https://github.com/GreyAssoc/cogai) |
| Open a support request | support@getcog.ai |

---

**Last updated:** 2026-06-06