# cog — public distribution

> A Go-native AI agent system you run yourself, built on an append-only,
> hash-linked event log. This is the public distribution surface: install
> instructions, the gear-authoring contract, and a
> [Releases page](https://github.com/GreyAssoc/cogai/releases) for the
> compiled binaries. The source lives in a private dev repo; you don't
> need it to run cog Free.

---

## The thing that makes cog different

Most agent frameworks keep a log of what the agent did. **In cog, the log
*is* what the agent did.**

Every action the runtime takes — every model call, tool invocation,
sub-agent dispatch and chain step — is recorded as a pair of events in one
strictly ordered, append-only, hash-linked log. Every view is then computed
*from* that log: your session, the audit trail, the redacted transcript,
the cost tally, the replay tape. Nothing is stored twice; nothing drifts
out of sync.

Three things fall out of that which most agent stacks structurally cannot
offer:

**Byte-exact replay.** Take a completed session, re-run it against its own
recorded model and tool responses, and it reproduces byte-for-byte —
*through the database*, not just in memory. That turns prompt changes and
model upgrades into a regression test: change a prompt, replay real
sessions, see exactly what moved.

**An audit trail you verify rather than trust.** Each event is
cryptographically bound to its predecessor, so one value commits the whole
history and any insertion, deletion, reorder or edit is detectable.

**Real runs become evaluation and training data.** Because the log is
complete and schematised, trajectories fold into fine-tuning-shaped
datasets, and one task can be diffed across several models.

**Why cog can do this and most tools can't:** it owns its agent loop
end-to-end — no vendor SDK, no MCP runtime, no shelling out to another
agent binary. Tools are typed Go functions running *in* the process, so
every effect crosses a boundary cog controls and can record. A stack whose
tools live in separate processes cannot replay byte-exact: the boundary
runs through something it doesn't own.


### Why there is no MCP runtime

Cog's tools are typed Go functions compiled into the binary, not
separate-process servers. That is the decision everything above rests on,
so it's worth explaining rather than asserting.

**What MCP gets right.** The Model Context Protocol is language-agnostic,
hot-pluggable, and has by far the largest tool ecosystem. If your priority
is breadth quickly, it is the correct choice and cog is not competing with
it on that axis.

**The structural gap.** MCP standardises *what the model may call*. It does
not standardise *what the callee may then do*. A server is an ordinary
process on your machine: it can read files, open sockets, and spawn shells.
The host says "call `search_files` with these arguments"; the server decides
what `search_files` actually does. There is no in-band way for a host to say
*"this tool may read `~/project` and nothing else"* and have that enforced.
You are trusting the server's author, and the protocol has no vocabulary for
distrusting them.

Three consequences follow from the architecture itself. They are not bugs to
be patched — they are what a process boundary means:

- **Permissions are coarse and advisory.** Consent is granted per tool, or
  per session, not per call with its actual arguments. Once a tool is
  approved, every later invocation inherits that approval whatever the
  arguments are.
- **The audit trail records intent, not effect.** The host can log the call
  and the value returned. What the server *did* — which files it touched,
  which hosts it contacted, which credentials it used — happened somewhere
  the host cannot observe. For a compliance reader that is the difference
  between "the model asked for X" and "X is what occurred."
- **Byte-exact replay is impossible.** You cannot tape an effect boundary
  that runs through a process you don't own. Everything in the section above
  — replay as a regression gate, a verifiable log, trajectories as
  evaluation data — requires that every effect cross a boundary the runtime
  controls.

**And four practical risks** that are *not* inherent to the protocol —
careful hosts and operators can mitigate all of them, and the spec has grown
to help — but which are common in real deployments and worth naming:

- **Installing a server is arbitrary code execution.** The usual flow runs
  third-party code with your user's privileges, unsandboxed. The tool you
  audited and the code that ships next week are not the same artefact.
- **Tool descriptions are model-visible text.** A server's own metadata
  enters the model's context, so a malicious or compromised server can carry
  prompt injection in its description — and can influence how the model uses
  *other* servers' tools sharing that context.
- **Definitions can change after you consented.** Approval is a moment;
  the served definition is a variable. Nothing structurally binds the tool
  you approved to the tool you later call.
- **Credentials spread out.** Servers commonly hold their own API tokens and
  OAuth grants, so every added server is another place secrets live and
  another independent blast radius.

**How cog answers this.** Not by being more careful — by removing the
boundary that creates the problem:

| The problem | Cog's answer |
|---|---|
| The server decides what a tool does | A gear is a Go function in this binary. No separate process, no install step that runs someone else's code. |
| Approval is per-tool, not per-call | Every gear declares a **permission envelope the engine enforces before it is invoked**, evaluated against the actual arguments of that call. |
| Tools can reach anything the process can | Gears **cannot reach raw filesystem, subprocess or network primitives at all.** Those live behind an internal `effect` package and are unexported at its boundary — the Go compiler refuses the import. |
| Enforcement drifts over time | That boundary is a **build gate**, not a review convention. A guard fails the build on any new raw import, and it currently reports **zero exceptions** across the whole gear surface. |
| The audit records intent, not effect | Every effect passes one dispatch chokepoint that writes an event **before** the call and another **after** it. The log records what happened, not what was requested — which is also precisely what makes replay work. |
| Tool definitions can shift under you | **All executable code is fixed at compile time** — cog loads no third-party code at runtime, ever. You can add declarative gears (YAML) without recompiling, but those load *data, not code*: each runs the same engine-owned, code-reviewed handler inside an envelope validated when the file loads. The inventory can grow; the code that executes it cannot. |
| Credentials spread across servers | One process, one credential surface, with secret-shaped strings scrubbed on the write path. |

There is deliberately **no general `bash` gear**, either. Shell access was
removed from the default surface: every subprocess a gear spawns now goes
through the audited chokepoint under an operator allowlist.

**What this costs you, plainly.** You cannot point cog at an arbitrary MCP
server and have it work. That is a real loss and we are not going to pretend
it isn't — the trade is **less ecosystem, more enforceable safety**. Two
things soften it: declarative gears let you add HTTP and webhook tools as
validated YAML against a [published schema](#authoring-declarative-gears),
no compiler needed; and an `mcp-to-gear` converter on the roadmap reads an
existing MCP server's tool list and generates auditable Go for you to review
and compile in. That is a build-time bridge, not a runtime one. **Cog will
not speak MCP at runtime** — doing so would forfeit every guarantee in the
section above.

### One capability set, two ways to drive it

The tool boundary above buys something beyond safety: because every
capability is a typed function inside one process, **composing capabilities
needs no second system.**

Most stacks that do real work end up running two. An agent framework for the
ambiguous parts, and a workflow tool — Zapier, n8n, Airflow, a pile of cron
and glue — for the repeatable ones. That means two sets of integrations to
build, two trust domains to reason about, two audit trails that don't line
up, and a seam between them where the credentials live.

Cog has one capability set and two ways to drive it:

| | **Agent loop** | **Chain** |
|---|---|---|
| Who decides the sequence | The model, at runtime | You, in YAML, ahead of time |
| Right for | Ambiguous work — classify, debug, draft, investigate | Repeatable work — the Monday-morning digest |
| Cost shape | N turns of reasoning | Exactly the model calls you placed |
| Started by | A message | Webhook, schedule, or an agent mid-conversation |

Both call **the same gears**. A gear you write once is available to the
model *and* as a step in a workflow — there is no integration layer to build
separately, secure separately, or audit separately.

Both run inside **the same enforcement boundary**. The chain executor is not
a privileged orchestrator sitting above the permission system: every step
runs at the gear's own tier, through the same dispatch chokepoint, under the
same envelope. A chain cannot reach anything the agent couldn't.

Both write to **the same event log**. So a workflow run is replayable and
verifiable in exactly the way a conversation is — same hash chain, same
byte-exact guarantee, one query to reconstruct a run. In a two-system setup
the automation half is usually the part with no meaningful audit trail at
all; here it's the same trail.

#### How gears link deterministically — and where reasoning fits

A chain links gears by **passing typed values**, not by handing the next
step a blob of text and hoping. Each step names a gear and declares its
input, and an earlier step's output is addressed **field by field** —
`{{ .extract.output.text }}`. The author wires named field to named
parameter, and a reference to a step that hasn't run is rejected when the
chain loads, not when it fires.

For an ordinary gear that's trivially deterministic: its output shape is
fixed by its own contract. **The only place a chain can drift is where you
need judgement — and that's exactly where the `reason` gear clamps it.**

`reason` is an ordinary gear that happens to invoke a model. You give it
instructions plus a `response_schema`, and its reply is **validated against
that schema before the step returns**. Non-conforming output doesn't
propagate — it's a step error. So the values arriving at the next gear are
typed even though a model produced them:

```
pdf_extract  →  reason  →  gmail_send
   text          { subject, summary, recipients }      typed parameters
                 ↑ shape fixed by response_schema
```

Read the middle step as a **shape adapter**: unstructured text in, a
validated object out, and `gmail_send` receives ordinary parameters with no
idea a model was involved. Nondeterminism is confined to the *wording
inside* those fields — it cannot change which fields exist, their types or
their constraints. Constrain a field with an `enum` and the routing decision
is genuinely closed even though a model made it.

That confinement is the point: **reasoning where you need it, typed plumbing
everywhere else** — about as close to deterministic outcomes as a system
with a model in it can honestly get. And you decide, field by field, which
parts are locked and which are free to be prose.

→ Worked example, the determinism dial, retry behaviour and resource bounds
in [Chains](#2-chains--declarative-workflows-where-every-node-can-think).

### Where this honestly stands

The substrate is mid-flight. What holds today versus what doesn't:

| | Status |
|---|---|
| **Byte-exact replay** | **Works today** — through database persistence, across all 8 providers, property-tested. One residual: a pathologically large *incompressible* session stores a verifiable bound marker rather than the full tape. |
| **Hash-linked, verifiable log** | **Works today.** Integrity-checked by default. It becomes **tamper-evident against someone holding database write access** only when you configure a signing key *and* keep the freshness anchor where that attacker can't reach it — realistic on managed Postgres, not on a single box. We don't use the stronger word for deployments that haven't earned it. |
| **Selective disclosure** | Inclusion and consistency proofs are **built** — proving one interaction is in the sealed log without revealing the rest — but not yet the committed structure. |
| **Replay & corpus tooling** | The guarantee holds; the **command-line tooling is built and tested but not yet released.** |
| **Determinism** | **Honest by design.** A published table ranks how reproducible each provider can be; a run that *can't* be reproduced is reported as such, not papered over. |

## Enforcement at the boundary

The tool boundary above is one half. The other half is what happens at
runtime once a call is permitted: **cog refuses rather than degrades.**

- **Budgets are compared as integers in micro-USD**, so floating-point
  drift can't slip a request past a spend ceiling. NaN or infinity in a
  budget column refuses the request instead of degrading to "no cap", and
  a per-user lock closes the race where two simultaneous requests both
  observe "under cap".
- **Policy lookups fail closed** in production posture — a database error
  refuses the request rather than waving it through.
- **A disabled gear is absent from the toolset**, not an error at call
  time, so the model never sees a capability it isn't allowed to use.
- **Path containment resolves symlinks**, `.git/` is refused in write
  gears, and the per-project skills directory is refused outright in
  default deployments because the agent's own `write` gear can reach it.

Cog has **not** had a third-party security audit, and there is no SOC 2 or
ISO 27001. The hardening above came from internal adversarial review —
run by us, on our own code — plus the design guards that fail the build.
Treat it as evidence of care, not as an external attestation.

## What you get

- **8 model providers** — Anthropic, OpenAI, Google Gemini, DeepSeek, xAI,
  Qwen, Moonshot, z.ai. BYO keys, switch per message, no token markup, and
  cog never restricts provider or model by tier.
- **~100 typed gears** — file I/O, code intelligence, git and build
  toolchains, sandboxed execution, web fetch and search, `.docx` /
  `.xlsx` / `.pdf` / `.pptx` read and write, Google Workspace, email,
  memory, cron and async tasks. Every one is callable both by the agent
  loop *and* as a step in a [chain](#2-chains--declarative-workflows-where-every-node-can-think).
  [Full catalogue below](#reference--the-full-catalogues).
- **37 built-in agents and 11 skills**, free on every tier — a flagship
  coding agent across 15 languages, research and writing specialists, and
  28 single-perspective code reviewers that Pro can run *as coordinated
  panels* with a chair and VETO authority.
- **Persistent pgvector memory** across restarts, cross-session and
  cross-project. Never gated.
- **Your data in your Postgres**, under a documented schema. Point Looker,
  Metabase or `psql` at it and never ask permission.
- **Local-first and BYO-keys** — cog runs on your hardware, you supply your
  own model keys, and the Free tier needs no licence file, no signup and no
  account.

> **On lock-in, precisely.** Cog is closed-source and paid tiers are gated
> by a signed licence, so this isn't open source and we won't pretend
> otherwise. What you do get is **data portability**: your traces live in
> your database under a schema we publish, you bring your own model keys
> with no markup, there's no hardware fingerprinting, and if cog disappeared
> tomorrow the binary you deployed keeps running and the SQL keeps
> answering.

---

## Install

### Guided setup via Docker (recommended)

If you have Docker installed, this is the fastest path. One
interactive container walks you through the same setup the native
installer does — Telegram bot token, allowed user IDs, your model
provider API key(s), default model, install location — and writes
`docker-compose.yml` + `.env` + `cog_mounts.yaml` to your current
directory:

```bash
mkdir cog && cd cog
docker run --rm -it -v $(pwd):/setup greyassoc/cog-installer:v0.4.0
docker compose up -d
```

Open Telegram, find your bot, send `/help`. The first run pulls
~120 MB (gateway image + bundled `pgvector` Postgres) and
typically completes in under 60 seconds on a fast connection.

The Discord variant of the gateway is a separate image and gets
wired in automatically by the installer if you supply a Discord
bot token:

```bash
docker pull greyassoc/cogai-discord:v0.4.0
```

All three images publish multi-arch (`linux/amd64` + `linux/arm64`).

> **Pin the version, don't track `latest`.** Every command here names an
> explicit tag on purpose. This page argues that a tool you audited should be
> the tool that runs; a floating tag gives away exactly that property. For
> production, go further and pin the digest —
> `greyassoc/cogai@sha256:<digest>` — which is immutable by construction.

### Hand-rolled compose (advanced)

If you want to skip the installer and write your own
`docker-compose.yml`, pull the gateway image and bundle it with
`pgvector/pgvector:pg16`. See [`DEPLOY.md`](https://github.com/GreyAssoc/cogai/blob/main/DEPLOY.md)
for the required env vars (Telegram token, allowed user IDs, at
least one model provider key, Postgres URL).

```bash
docker pull greyassoc/cogai:v0.4.0
```

### Native installer (no Docker prerequisite)

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

## Tier matrix — what each tier unlocks

The principle, from `TIERS.md §1`: **never gate things that make cog useful as a coding agent.** Memory, plan mode, cron, code intelligence, every model provider, every built-in agent, every built-in gear are always Free. Paid tiers gate *coordinated execution*, *declarative automation* and *governance* — features that compound with scale, not features needed to ship one project.

| Capability | **Free** *(shipping)* | **Pro** *(shipping)* | **Teams** *(planned)* |
|---|---|---|---|
| **Seats** | 1 | 1 | unlimited |
| **All 8 model providers** | ✓ | ✓ | ✓ |
| **All models, no token markup** | ✓ | ✓ | ✓ |
| **All ~100 built-in gears** | ✓ | ✓ | ✓ |
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
| **Custom gears beyond the built-ins** | 3 | **unlimited** | **unlimited** |
| **Custom agents beyond the built-ins** | 3 | **unlimited** | **unlimited** |
| **Custom skills beyond the built-ins** | 3 | **unlimited** | **unlimited** |
| **Council (4-agent parallel + chair + refinement)** | ✗ | **✓** | **✓** |
| **Orchestrator (`subagent` gear)** | ✗ | **✓** | **✓** |
| **Bundled agent teams** (go-dev-team, fe-sec-team, …) | ✗ | **✓** | **✓** |
| **Custom teams** (your own members + chair + VETO) | ✗ | **✓** | **✓** |
| **Chains** (declarative YAML workflows + the `reason` gear) | ✗ | **✓** | **✓** |
| **Token usage reporting** (total + **per-agent**) | ✗ | **self** | **org-wide** |
| **Self-imposed spend cap** | ✗ | **✓** | **✓** |
| **Audit / governance dashboards** | ✗ | self-only | **full multi-user** |
| **Per-user policy mutation (budget caps, allowed providers, quotas)** | ✗ | self-only | **admin over users** |
| **OIDC SSO** | ✗ | ✗ | **✓** |
| **GDPR Art. 15 subject-access endpoint** | ✗ | ✗ | **✓** |
| **Reporting endpoints (usage / violations / cost outliers / transparency)** | ✗ | ✗ | **✓** |
| **Trace retention** | 30 days | 365 days | configurable |
| **Channels** | Telegram + Discord | + WhatsApp, HTTP API, operator dashboard | + on-prem connector |
| **Source-access rider (audit the engine under NDA)** | ✗ | ✗ | **✓** |
| **No phone-home** | ✓ | daily check-in (licence_id only) | configurable: online or fully offline |
| **Licence file required?** | no | yes (Ed25519-signed) | yes (Ed25519-signed, org-bound) |

Tier is resolved **only** from an Ed25519-signed licence file, verified locally against a public key embedded in the binary. No phone-home for Free, no signup, no account. The tier policies are compiled into the binary, so there is no config file to edit your way into a paid feature.

### Free is a credible standalone product

Free isn't a teaser. You get a real coding agent that holds its own on the coder-review-coder loop, **all** model providers (BYO keys, no token markup), persistent memory across restarts, cron, plan mode, 37 built-in agents, ~100 typed gears, 11 skills, and 30-day forensic-grade audit retention. Everything you need to ship one project end-to-end, perpetually, without giving us an email address.

---

# Pro — the single-seat power tier

*Shipping since v0.4.0, with its gates enforced at the trust boundary: Free gets a hard refusal, not a warning.*

Free gives you one very good agent. **Pro gives you a coordinated system of them, a way to automate it, and the numbers to run it on.**

## 1. Parallel agent coordination

The bright line between Free and Pro isn't *which* agents you can reach — it's whether they can run **together, under a protocol**.

| Capability | Free | Pro |
|---|---|---|
| Invoke `@go-security` on its own, read its findings | ✓ | ✓ |
| Run `@go-dev-team` — three agents in parallel, chair synthesises, VETO honoured | ✗ | ✓ |
| Compose your own team (members + chair + protocol + who holds VETO) | ✗ | ✓ |
| Dispatch a sub-agent with its own context from inside a session | ✗ | ✓ |
| Convene a 4-person council on a hard question | ✗ | ✓ |

### Council — four minds, one answer

`/council <question>` runs your question through **four isolated councillor agents**, each with a deliberately different epistemic stance, then a **chair** that reads all four and produces the answer you see.

| Councillor | The stance it argues from |
|---|---|
| **First-principles** | Derives from foundations, states its assumptions, and says so when the question rests on a faulty premise. Doesn't defer to received wisdom. |
| **Empiricist** | Separates what is *known* (citable) from what is *believed*, quantifies uncertainty, and asks "how would I know if I were wrong?" before committing. |
| **Devil's advocate** | The loyal opposition. Finds the strongest argument *against* the obvious answer and surfaces what you might not want to hear. Forensic, never gratuitous. |
| **Domain pragmatist** | The long-experienced practitioner. Weighs real constraints and common pitfalls; allergic to elegant theories that have already been tried and failed. |

Councillors are peers, not ranked. Each runs as an isolated agent with no shared state, so their errors are less correlated than four samples from one prompt. An optional refinement round lets each councillor see the others' first-pass opinions and revise.

**What it costs, stated plainly:** roughly **5× the tokens** of a normal turn, and the latency of the slowest councillor plus the chair. The honest claim is *"fewer obvious mistakes on hard questions"* — not maximum accuracy. There is deliberately **no fact oracle**: four councillors who disagree give you a more honestly hedged answer, not certainty. Reserve it for genuinely high-consequence questions; it is emphatically not the default. The full deliberation lands in your trace store, so you can read what each councillor actually said rather than trusting the synthesis.

### Agent teams — the review panel, coordinated

A team is a **coordinated multi-agent pattern**: several specialists run on the same problem, a chair synthesises, and a consensus protocol respects **VETO authority**. Eight teams ship pre-bundled:

| Team | Members | For |
|---|---|---|
| `go-dev-team` | go-purist, go-pragmatist, go-pessimist | Go code quality |
| `go-sec-team` | go-security **[VETO]**, go-hacker | Go security |
| `fe-dev-team` | fe-purist, fe-pragmatist, fe-pessimist | Frontend quality |
| `fe-sec-team` | fe-security **[VETO]**, fe-hacker | Frontend security |
| `mobile-review` | kotlin-purist, mobile-security **[VETO]**, mobile-hacker | Kotlin / Android |
| `design` | security, compliance, data-integrity (all **VETO**), performance, error-handling | Design review |
| `integration` | api-guardian, cross-language-security **[VETO]**, data-flow, failure-modes | Integration review |
| `research` | ui-ux-researcher | UI/UX |

Three coordination modes: **parallel** (all members on the same prompt, chair synthesises), **sequential** (each member sees the prior one's output), and **consensus** (parallel plus a negotiation round when members disagree, with Tier-1 VETO able to block the verdict outright).

Custom teams are a YAML file naming members, chair, mode and VETO holders — composable from any built-in *or* custom agent.

### Orchestrator

The `subagent` gear dispatches a sub-agent **with its own context** from inside a running session — a genuine second agent loop, not a persona switch. Its token spend is attributed back to it by name (see §4). `dispatch_agent`, the lighter hand-off to a built-in specialist, stays free on every tier; it's the multi-context pattern that's Pro.

## 2. Chains — declarative workflows where every node can think

The agent loop is the right primitive for **ambiguous** work: classify this
email, debug this failure, draft this reply. It is the wrong primitive for
deterministic glue. *"Every Monday at 09:00, summarise last week's tickets
and post the digest"* shouldn't require you to sit in a chat session, and it
shouldn't cost a fresh round of model reasoning to decide what it already
knows it's doing.

**Chains** are YAML workflow definitions that link gears together with
templated data flow. Same gears, same permission model, same audit log — you
just decide the sequence instead of letting the model decide it.

### How gears and chains fit together

This is the part worth understanding, because it's why chains aren't a
bolted-on automation product.

**A gear is one typed capability.** One call, one schema, one permission
envelope the engine enforces before it runs.

**A chain is an ordered composition of gears.** Each *step* is one named
invocation of one gear, and its output becomes available to the steps after
it — the wiring is covered in detail
[below](#how-the-output-of-one-gear-becomes-the-input-of-the-next).

**The `reason` gear is the hinge.** It is *a gear like any other* — it just
happens to invoke a model. Give it `{instructions, data, response_schema}`
and it returns JSON matching your schema. That single design choice is what
makes LLM judgement a **node in the graph** rather than the thing that
drives the graph. You place intelligence exactly where the problem needs it
and nowhere else.

**Chains introduce no new trust tier**, which is what makes the
[one-capability-set property](#one-capability-set-two-ways-to-drive-it)
hold in practice. The chain executor is not a privileged orchestrator
sitting above the permission system: every gear a chain step calls runs at
its own existing tier, under its own envelope, through the same dispatch
chokepoint. A chain cannot smuggle a capability into a context that wasn't
already allowed it — and every gear call inside a run is stamped with the
run's id, so one SQL query reconstructs the whole thing.

### How the output of one gear becomes the input of the next

This is the mechanism that makes a chain deterministic, so it's worth
spelling out concretely.

Each step names a gear and supplies an `input:` block. That block is a
**template rendered at dispatch time** against the run's namespace, and
every earlier step's output is addressable in it **field by field**:

```
{{ .trigger.body.<field> }}        what fired the run
{{ .<step-name>.output.<field> }}  an earlier step's output, one field of it
{{ .secret.<NAME> }}               resolved at render time, never written to a trace
```

So step 3 doesn't receive "whatever step 2 produced" as an opaque blob — the
author wires **named field to named parameter**, and the loader refuses at
load time if a step references one that hasn't run yet. The wiring is
explicit, and a mistake in it is a startup error rather than a 3 a.m. one.

For an ordinary gear, the shape of `output` is fixed by that gear's own
contract: `pdf_extract` returns text, `gmail_send` returns a send result.
Deterministic in, deterministic out.

**The interesting case is where you need judgement.** That's the one place a
chain could go non-deterministic — and it's exactly where the `reason` gear
clamps it. `response_schema` is **required by default**, and the model's
reply is validated against it *before the step returns*. Non-conforming
output does not propagate; it's a step error. So the fields flowing into the
next gear have a guaranteed shape even though a model produced them.

Put together, it looks like this — a real three-step chain, messy input to
typed decision to deterministic action:

```yaml
steps:
  - name: extract                       # deterministic gear
    gear: pdf_extract
    input:
      path: "{{ .trigger.body.pdf_path }}"

  - name: summarise                      # the reasoning step
    gear: reason
    input:
      instructions: |
        Summarise this report for an executive audience.
      data: "{{ .extract.output.text }}"      # <- output of step 1 in
      response_schema:                        # <- the clamp
        type: object
        required: [subject, summary, recipients]
        properties:
          subject:    { type: string, minLength: 5, maxLength: 200 }
          summary:    { type: string, minLength: 50 }
          recipients:
            type: array
            minItems: 1
            items: { type: string, format: email }
      gear_subset: []                         # pure judgement, no tool use

  - name: send                           # deterministic gear again
    gear: gmail_send
    input:
      to:      "{{ .summarise.output.recipients }}"   # <- typed fields out
      subject: "{{ .summarise.output.subject }}"
      body:    "{{ .summarise.output.summary }}"
```

Read the middle step as a **shape adapter**. Unstructured PDF text goes in;
an object with a validated `subject`, `summary` and `recipients` comes out;
`gmail_send` receives ordinary typed parameters and has no idea a model was
involved. The nondeterminism is confined to the *wording inside those
fields*. It cannot change which fields exist, their types, or their
constraints — `recipients` will be a non-empty array of things shaped like
email addresses or the step fails and the value never reaches `gmail_send`.

That confinement is the whole idea: **reasoning where you need it, typed
plumbing everywhere else.** Constrain a field with an `enum` and the model
*cannot* return a value outside it, so a routing decision is genuinely
closed even though a model made it. Leave a field a bare `string` and it
varies freely. You choose, field by field, which parts of the outcome are
locked and which are allowed to be prose.

**When a provider fails to honour the schema**, the gear retries with a
budget scaled to how well that provider supports structured output — none
for providers with native JSON Schema, more for weaker ones — and every run
records whether the model conformed first time, how many retries it took,
and what the violations were. Schema failures are visible operational data,
not a silent fallback to something unshaped.

> **One honest limit.** Cog does not statically prove that step 2's output
> schema satisfies step 3's declared input — there's no whole-chain type
> check at load time. What is checked at load is that every gear exists and
> that no step reads a step that hasn't run; what is enforced at run time is
> the reasoning step's own output schema. In practice that covers the case
> that actually breaks workflows — a model returning something unexpected —
> but wiring the wrong field to the wrong parameter is still an authoring
> error you find by running the chain.

### Determinism is a dial you set, not a property you hope for

Most workflow tools with an "AI step" hand you a black box. Cog treats
determinism as three separate questions, and a chain answers each one
according to how *you* built it:

| Dimension | The question | Why you care |
|---|---|---|
| **Path** | Does the same input run the same sequence of steps? | Cost predictability, failure-mode reasoning |
| **Output** | Do the same inputs produce the same bytes? | Regression tests, idempotency, replay |
| **Outcome** | Do the same inputs produce the same *business effect*? | The one anyone actually feels |

| Chain shape | Path | Output | Outcome |
|---|---|---|---|
| No `reason` steps at all | Fully | Deterministic bar external services | Fully |
| `reason` + strict schema, no tool use | Fully | Shape locked, wording varies | Effectively — decisions are schema-bound |
| `reason` + strict schema + nested tool use | Steps fixed, nested calls vary | Shape locked, wording varies | Mostly |
| `reason` with schema opted out | Fully | Highly variable | Not guaranteed |

**The schema is the knob.** A field declared `enum: [escalate, close, reply]`
*cannot* return anything else — validation refuses to propagate a
non-conforming value downstream. A bare `type: string` can return whatever
the model writes. So you lock the fields that drive downstream actions and
let the human-readable narrative vary.

Concretely: three runs of an unschematised classifier might return
`{"decision":"escalate"}`, then unparseable prose, then
`{"verdict":"ESCALATE"}` — breaking downstream steps three different ways.
Bind the schema and the routing is locked to *3 decisions × 3 teams ×
5 priorities = 45 possible outcomes*, forever, while the rationale wording
stays free.

**And you write that schema once.** Providers vary enormously in schema
support — native JSON Schema on some, tool-shaped schemas on others,
`json_object`-only or prompt-only on the rest. The `reason` gear looks up the
active provider and emits the right wire format for it, with automatic
retry when a weaker provider drifts. You author one `response_schema`; the
gear handles the per-provider reality.

### Triggers, and how a chain gets started

- **Webhook** — HMAC-signed by default, constant-time verified, and a failed
  auth returns 401 with no body. Compatible with how GitHub, Stripe and
  Shopify already sign.
- **Schedule** — cron, using the same scheduler as everything else.
- **Agent dispatch** — an agent *inside a conversation* can fire a chain and
  keep going. This is where the two modes compose: reason about a messy
  situation in the agent loop, then hand the repeatable part to a chain that
  does it the same way every time.

### Bounded by construction

Because a chain is operator-authored YAML that can run unattended, the
limits are defaults rather than options:

| Bound | Default | On hit |
|---|---|---|
| Wall-clock timeout | 300s | Run cancelled, remaining steps skipped, recorded as timed out |
| Gear-call cap | 50 | The step that would exceed it is refused *before* dispatch |
| Cost cap | $1 per run | Same — refused before spending, not detected after |

An operator ceiling caps what any individual chain may ask for. Template
access to environment variables is **empty by default** — a chain sees no
env at all unless you allowlist specific names — because leaking a provider
key into a trace through a template would be unrecoverable. Secrets resolve
through the same path gears use and are never serialised into traces. A
`reason` step inherits the caller's allowed gears and can be narrowed
further per step; an empty tool list means no tool use at all, which is the
safest configuration and worth using whenever the step is pure judgement.

Chains execute **in-process** — outputs move through memory, not HTTP. Scale
horizontally by running more cog instances against the same Postgres.

## 3. Unlimited extensibility

Free lets you author 3 custom gears, 3 agents and 3 skills on top of the built-ins. **Pro removes all three ceilings.** This is the tier where cog conforms to your workflow rather than the reverse.

| | What it is | Free | Pro |
|---|---|---|---|
| **Custom gears** | Declared HTTP (Tier 1 YAML against the public `cog_gear: v1` schema), declared webhook (Tier 2), or `mcp-to-gear` converter output | 3 | unlimited |
| **Custom agents** | A system prompt + tool list + persona overlay, registered under a name | 3 | unlimited |
| **Custom skills** | A YAML declaring trigger regexes, required gears, and the procedural fragment injected when the matcher fires | 3 | unlimited |
| **Custom teams** | Members + chair + coordination mode + VETO holders | ✗ | ✓ |

Quotas are enforced at **registration**, not at runtime, and the behaviour on tier change is deliberately non-destructive: registrations beyond a lower tier's cap are preserved and marked inactive rather than deleted, and re-activate in registration order on upgrade. **Downgrading never eats your work.**

## 4. Token usage reporting — see which agent spent your money

`agent_id` is threaded through sub-agent dispatch onto every token-bearing event, which makes something no tier had before possible: **per-agent cost attribution**.

When a council run costs 5× a normal turn, you can see *which councillor* spent it. When an agent team reviews a diff, each member reports its own tokens. The orchestrator's sub-agents are itemised by name.

Three ways in:

- **`/usage [Nd]` in chat** — total tokens and cost for the period (default 30 days), then a per-agent breakdown sorted by cost.
- **`cog usage -user <id> [-days N]`** — the same aggregation as a CLI table.
- **An optional Metabase dashboard** — a single-file compose add-on with six documented, copy-paste questions: total, per-agent (the headline), per-model, daily trend, and — at Teams — per-user and per-user × agent. The setup guide includes the read-only-role `GRANT` and payload-privacy guidance.

**Strictly self-scoped.** A Pro user only ever sees their own spend.

**Self-audit** gives you your own usage, failures and sessions over the same trace store — and because that store is *your* Postgres with a documented schema, you can point Looker, Tableau or plain `psql` at it and never ask permission.

> **Honest caveat:** `cost_usd` is an estimate from per-model price tables that change. Treat it as best-effort operational signal; the authoritative invoice is your provider's dashboard.

## 5. The spend cap — a safety primitive, not a throttle

Pro sets its own ceiling: *"don't let me burn more than £100 this month."* Free users don't see the cap UI at all — they own their provider bill directly, and that bill is its own protection.

The enforcement underneath is engine-side and the same at every tier: budgets are compared as **integers in micro-USD** so floating-point drift can't slip a request past the ceiling; NaN or infinity in a budget column refuses the request rather than silently degrading to "no cap"; and a per-user lock closes the race where two simultaneous requests both see "under cap" and collectively overshoot. **Soft warning at 80%, hard refusal at 100%.**

## 6. Interfaces, channels and retention

- **WhatsApp** — Pro is the first tier with it. Meta charges per conversation above their free allowance, so the cost is wrapped into the subscription rather than pushed onto Free.
- **HTTP API** — Bearer-token authenticated, per-source-IP rate limited *before* the token compare so brute-force attempts don't exercise the auth path at line rate.
- **Operator dashboard** — a server-rendered view of your agents, teams and recent dispatches, embedded in the gateway binary. Vanilla HTML5 + CSS3, no framework.
- **Chain webhook endpoint** — the HTTP trigger surface for chains.
- **`cog` CLI** — usage reporting and licence management.
- **365-day trace retention**, up from 30.

> **Built vs. planned, stated plainly.** The three shipping channel hosts are **Telegram, Discord and WhatsApp**. Web chat, an IDE host and the on-prem connector are on the roadmap — their gates exist, their host binaries don't yet. They are not in your download today.

---

# Teams — multi-user governance

**Status, up front:** the Teams tier is **planned, not yet purchasable.** The machinery below is built — the admin plane ships its admin UI, OIDC, signed-licence enforcement, and reporting and transparency endpoints. What isn't open yet is licence issuance for the tier, and there is no hosted SaaS: **self-hosting is the answer today.**

Everything in Pro, **for every seat**, plus the whole admin plane.

## The structural idea: the admin plane is not in the request path

This is the design decision the whole tier rests on.

```
┌──────────────────────────────────────────────────────┐
│  admin plane                                         │
│  • identity / SSO bridge   • policy + budget         │
│  • reporting + admin UI    • licence validation      │
└──────────────────────┬───────────────────────────────┘
                       │ reads + writes config
                       ▼
        Postgres + pgvector  (shared trace store)
                       ▲
                       │ writes traces, reads policy
┌──────────────────────┴───────────────────────────────┐
│  sub-cogs — one process per employee                 │
│  • talks to the model provider DIRECTLY              │
│  • policy fetched at session start, cached with TTL  │
│  • every trace row tagged with user_id               │
└──────────────────────────────────────────────────────┘
```

**Policy is mutated in the admin plane but enforced in the engine.** An outage of the admin plane must not stop your people working — it stops *changes to policy* and *new reports*, nothing else. Sub-cogs cache policy with a short TTL and keep going.

**One process per employee**, not tenants in one process, because per-process isolation is the simplest defensible boundary: one person's agent loop can't bleed memory or sessions into another's, a compromised sub-cog doesn't compromise everyone, and different people can run different versions during a rollout.

## Per-user policy — the dials an admin actually turns

| Control | Effect |
|---|---|
| Monthly budget cap (USD) | Hard refusal at 100%, warning trace at 80% |
| Allowed providers / model allowlist | Request refused on a disallowed model |
| File read / write roots | Engine-enforced path containment |
| Bash allowlist / disable bash | Unset = engine default, empty = disabled outright |
| Cron enable / disable | Per-user scheduler access |
| Council enable + daily cap + concurrent cap | Bounds the 5×-cost feature per person |

Policies resolve **per-user and per-group, narrowest winning**, and the API returns the *effective* policy with its source so an admin can see why someone has the limits they have. Every change is bound to the authenticated admin's email for the audit trail.

**Where enforcement actually happens:** static fields (file roots, bash allowlist, excluded gears) resolve when the agent backend is constructed; dynamic fields (model allowlist, budget) are checked on **every request**. A disabled gear is simply *absent from the toolset* rather than failing at call time. In production posture, a policy-lookup database error **refuses the request** rather than failing open.

## Identity

- **OIDC SSO** — RS256 + JWKS, with single-flighted refresh so unauthenticated callers can't amplify into parallel fetches, an SSRF-guarded fetch client, a 2048-bit minimum on RSA keys, and clock-skew-tolerant claim validation. Entra, Okta and Google all work.
- **User IDs are opaque to the engine** — email, UUID, employee ID or SAML name identifier, your choice. The engine keys sessions and memory by it and never parses it. Identity is a *host* responsibility.
- **Sessions** carry a server-side opaque identifier rather than a literal bearer token in a cookie, with the CSRF token bound to the session.

## Reporting, org-wide

| Endpoint | What it gives you |
|---|---|
| `/api/usage?group_by=user\|gear\|model\|agent` | Pivot tables across every user — including the same per-agent token attribution Pro sees for itself |
| `/api/users` | Everyone, with last-active and month-to-date spend |
| `/api/users/{id}` | One person: sessions, gear breakdown, plan-completion rate, recent failures |
| `/api/violations` | Budget warnings, failures, permission denials, off-hours sessions, fetches to non-allowlisted domains |
| `/api/sessions/{id}` | The full trace timeline for one session |
| `/api/reports/{name}` | Six saved structural-signal queries |

The admin UI renders the same data as server-rendered HTML: a spend and failure dashboard, users, per-user drill-down, the policy pane with near-cap and over-cap rows highlighted, fact counts by category, and failures grouped by class and provider.

**The six structural-signal queries** — top spenders, off-hours activity, sessions touching no project files, cost outliers (>2σ above the mean), fetch destinations, and gears used by only one person.

> **What this deliberately is not.** Cog ships **no automated personal-use classifier**. The queries surface signals; a human interprets them. That's a product decision, and it isn't a configuration knob.

## Data access is compartmented

Four database roles, so the admin plane isn't a single trust blob: sub-cogs get **INSERT-only** on traces and read on policy; admins get full read/write on the admin schema; analysts get **read-only** on traces and can run reports; and employees get **read-only on their own rows**, for transparency.

## Verifying the audit trail itself

Two endpoints, and the distinction between them matters:

- One verifies that the **event hash chain** is intact, and labels itself as such so it can't be mistaken for the stronger result. It counts rows it *couldn't* check rather than passing a session that verified nothing.
- One verifies the **cryptographically signed trajectory head** using a **public key only** — no signing secret ever sits on the audit host.

The signed endpoint returns **"not configured" rather than a false "verified"** when no verify key is present, and reports rollback detection as *unavailable* when the audit host structurally can't reach the anchor that would provide it. **Refusing to answer is treated as better than answering wrongly.**

## The compliance surface

Cog in a workplace is a monitoring system, and the product treats that as a legal obligation rather than a feature list.

- **GDPR Art. 15** — a subject-access endpoint returns what the system holds about a person, in one call. Self is always allowed; a non-admin requesting someone else is refused. **Art. 17** erasure is a first-class endpoint too.
- **Lawful-basis notice as a first-class concept** — an admin-editable record, a web form, an API, and sub-cogs that display it on first run. It must be set **before any user is onboarded**: deploying without notice is a **breach of the licence**, not merely discouraged.
- **A published proportionality bound** — cog records prompts and gear activity from *inside the agent session*. It does **not** record keystrokes, screen, or anything else you do on your machine. That bound is product policy, not a setting.
- **Redaction on the write path** — secret-shaped strings (provider API keys, GitHub tokens, AWS credentials, JWTs, bearer tokens, database passwords, bot tokens) are scrubbed before anything is stored.
- **Configurable retention per row class**, with sensible defaults.

## Deployment topologies

| Shape | Fit |
|---|---|
| **Single-operator pilot** | One gateway + one Postgres + Telegram. What the project itself runs daily. Free is sufficient. |
| **Multi-user self-hosted** | Gateway + Postgres + admin plane with a signed licence, OIDC bridge, production posture behind a TLS-terminating proxy |
| **Air-gapped** | Self-hosted admin plane + sub-cogs + a local model (Ollama / vLLM with an Anthropic-format adapter). Licence delivered on signed media. **Zero network egress.** |
| **Hosted SaaS** | **Not implemented.** Self-host today. |

## What an enterprise should also know

- **Your data is in your Postgres**, under a documented schema. Wire your own BI over it; the admin plane is one consumer, not the only one.
- **If cog disappears**, your traces and the schema remain, the binary you deployed keeps working, and you can keep querying by SQL indefinitely. What stops is updates, support and new licences.
- **If our licence server is down**, paid binaries honour their cached expiry until it lapses, with a grace window on the daily check-in so a transient outage can't trip a spurious downgrade.
- **No phone-home for Free or air-gapped deployments.** Online paid tiers check in once daily carrying a licence id, an instance id and a version — no telemetry, no prompts, no user data.
- **No hardware fingerprinting.** Licences bind to organisation, email domain and seat count — not to a machine.
- **No obfuscation.** Enterprises auditing a tool that audits their staff need to read it; a source-access rider makes that contractual. The security argument is legal, not technical.
- **No SOC 2 / ISO 27001 yet.** The evidence base is there — audit trail, RBAC, retention controls, lawful-basis notice — but the audit hasn't been done. It gets done when a customer is paying for the resulting paperwork.

Pricing for Pro / Teams is deliberately deferred — see [getcog.ai/pricing](https://getcog.ai/pricing). The Free tier is perpetual, single-seat, BYO-keys; no signup or licence file required.

---

## Reference — the full catalogues

Everything below ships in the binary and is available on **every tier, including Free**. Listed for reference; you don't need to read it to get started.

### Agents (37)

All 37 agents below are bundled in the binary and available **free for every tier**. Invoke with `@<name>` in your bot, or set the agent in your session config. Paid tiers don't gate individual agent access — Pro+ gates *coordinated* execution (Council, orchestrator, agent teams).

#### General-purpose (9)

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

#### Code review — cross-language (10)

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

#### Code review — Go (5)

| `@name` | Concern | Tier |
|---|---|---|
| `go-purist` | Idiomatic Go; would Rob Pike approve? | 3 |
| `go-pragmatist` | Ship-vs-perfect tradeoffs in Go | 3 |
| `go-pessimist` | Race conditions, resource leaks, what can fail | 2 |
| `go-security` | Go-specific vulns, panic propagation, supply chain | 1 VETO |
| `go-hacker` | Adversarial probing of Go code | 2 |

#### Code review — Frontend (6)

| `@name` | Concern | Tier |
|---|---|---|
| `fe-purist` | Semantic HTML, vanilla ES2020+ correctness | 3 |
| `fe-pragmatist` | Browser-compat vs. cleanliness | 3 |
| `fe-pessimist` | Async edge cases, event-handler leaks, DOM resilience | 2 |
| `fe-security` | XSS, CSRF, CSP, supply chain | 1 VETO |
| `fe-hacker` | Browser-side adversarial probing | 2 |
| `fe-functional` | Functional correctness, event flow, state mutation | 3 |

#### Code review — Mobile / Kotlin (4)

| `@name` | Concern | Tier |
|---|---|---|
| `kotlin-purist` | Idiomatic Kotlin, structured concurrency | 3 |
| `mobile-security` | Android security model, permission abuse, supply chain | 1 VETO |
| `mobile-hacker` | Adversarial probing of mobile code | 2 |
| `android-mobile-coding` | Jetpack Compose, Hilt, Room patterns | 3 |

#### Code review — Integration (3)

| `@name` | Concern | Tier |
|---|---|---|
| `api-guardian` | API contract stability, version compatibility | 2 |
| `data-flow` | End-to-end data flow correctness | 2 |
| `failure-modes` | Failure injection, recovery paths | 2 |

#### Research helper (1)

| `@name` | Concern | Tier |
|---|---|---|
| `ui-ux-researcher` | UI/UX patterns, accessibility | 3 |

#### The coder-review-coder loop

The canonical Free-tier coding workflow. Reviewers don't have to be ganged up into a team to be useful — a single review pass between coder rounds catches most issues.

```
@cog-coder       writes / fixes code
@<review-agent>  critiques (any single agent, e.g. @go-security)
@cog-coder       revises based on the critique
@<other-agent>   (optional) second-pass review
```

This loop ships as a CI-tested regression suite. The full agent reference is the tables above — every one of the 37 is in the binary on every tier.

### Skills (11)

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

#### Storage layers (highest precedence first)

```
<workdir>/.cog/skills/*.yaml   — per-project overrides (refused if inside a FileWriteRoot)
~/.cog/skills/*.yaml           — operator-authored (the recommended layer)
engine/skills/builtin/*.yaml   — the 11 built-ins, embedded into the binary
```

Same-name skills in a higher layer fully replace lower layers — operators override a built-in by dropping `~/.cog/skills/git_commit_workflow.yaml` rather than patching the binary. Name normalisation (TrimSpace + ToLower) means a near-alias like "Create_Skill " still resolves to the built-in `create_skill` key. The `<workdir>/.cog/skills/` layer is refused in default deployments because the agent's `write` gear can reach it — keep custom skills in `~/.cog/skills/` on the host.

### Gears (~100)

Gears are typed Go functions the agent dispatches as tools — and the same
gears are the building blocks of [chains](#2-chains--declarative-workflows-where-every-node-can-think),
so anything here can be composed into a declarative workflow without writing
an integration layer. **No gear is tier-gated** — every one in the binary is available on every tier including Free. Pro+ unlocks the *quota for adding your own* (Free 3, Pro unlimited) plus the orchestration gears.

> Looking for `bash`? There isn't one — see [why there is no MCP runtime](#why-there-is-no-mcp-runtime) for the reasoning. Sandboxed execution is available via `code_exec` / `sandbox_exec`.

#### File I/O & search

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

#### Shell, build, VCS

| Gear | Purpose |
|---|---|
| `git` | Structured `git status` / `git diff` / common subcommands. No shell pipe risk. |
| `code_build` | Project's build command. Returns exit, duration, captured excerpt. |
| `code_test` | Project's test suite. |
| `code_status` | Aggregate of git status + last build/test status. |
| `code_diff` | Pending diff vs HEAD or a named ref. |
| `code_lint` | Run the configured linter. Returns structured findings when parseable. |
| `code_format` | Run the configured formatter on a file. |

#### Code intelligence (cog's static analysis index)

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
| `sandbox_exec` | Run a command inside an OS-level sandbox (nsjail / bubblewrap where present), under the operator's program allowlist. |

#### Web

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

#### Media + document extraction & generation

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

#### Google Workspace (BYO OAuth)

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

#### Memory + state

| Gear | Purpose |
|---|---|
| `remember` | Save a persona fact to pgvector-backed memory (cross-session, cross-project). |
| `forget` | Remove a previously-saved fact. |

#### Cron + orchestration

| Gear | Purpose | Tier |
|---|---|---|
| `cron_schedule` | List, add, remove, enable / disable, or run cron jobs. | Free |
| `dispatch_agent` | Hand off the turn to a specialist built-in agent (e.g. `@researcher`). | Free |
| `subagent` | Run a sub-cog with a separate context window for a focused subtask. | **Pro+** |
| `task_start` / `task_status` / `task_result` / `task_update` / `task_list` / `task_cancel` | Async task management — fire-and-forget long-running gears that survive turn boundaries. | Free |

#### Time

| Gear | Purpose |
|---|---|
| `date_time` | Current time, timezone, date arithmetic — the agent should never guess "today's date". |

---

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
