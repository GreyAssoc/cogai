# cogai — Cog Public Gear Schema & SDK

> Public artefacts for authoring **declarative gears** that run on
> Cog (the AI agent runtime). Cog itself is closed-source proprietary
> software; the *contract* by which third-party gears are submitted
> and validated is published here so anyone can author a gear without
> needing access to the engine source.
>
> This repository is permissively licensed (MIT) because nothing here
> is commercially load-bearing. The runtime that *interprets* gears
> against the contract is the proprietary part.

---

## What this repo contains

| Path | Purpose |
|---|---|
| `cog_gear.v1.schema.yaml` | Canonical meta-schema for the `cog_gear: v1` declaration format, in YAML for readability. |
| `cog_gear.v1.schema.json` | The same meta-schema as JSON Schema — used by editor extensions (VS Code YAML, IntelliJ) to provide completion and inline validation. |
| `cog-gear-lint/` | A Go-based CLI that validates a gear declaration against the schema. Run before submitting; CI integrators run it on PRs. |
| `examples/` | Reference gear declarations for Notion, Stripe, GitHub, Linear. Copy + adapt. |
| `docs/AUTHORING.md` | How to write a Tier 1 (HTTP) or Tier 2 (webhook) gear from scratch. |
| `LICENSE` | MIT. |

---

## What gears are

A **gear** is a typed tool that the cog agent loop can dispatch.
Some gears are compiled into the cog binary (Tier 0); others are
declared in YAML and loaded at runtime (Tier 1 and Tier 2). The
declarative tiers are what this repo enables.

| Tier | Authoring | Envelope (what the gear may do) |
|---|---|---|
| **0 — Native** | Go source, compiled into the cog binary | Anything: file I/O, subprocess, network, custom permissions |
| **1 — Declared HTTP** | YAML/JSON, loaded at startup | Network to one or more declared hosts; **no** subprocess; **no** file I/O |
| **2 — Declared Webhook** | YAML/JSON with `external_workflow:` admission | Same as Tier 1, plus visible degradation indicating the gear may transit Zapier/n8n/Make-style platforms |

This repo's schema covers Tiers 1 and 2 — what you can author
yourself without filing a feature request against the cog engine.

---

## Quick start

### Write a gear

```yaml
# my_notion_gear.yaml
cog_gear: v1
tier: http
name: notion_create_page
description: Create a new Notion page in a known database.
endpoint:
  method: POST
  url: "https://api.notion.com/v1/pages"
  headers:
    Authorization: "Bearer ${secret:NOTION_TOKEN}"
    Notion-Version: "2022-06-28"
    Content-Type: application/json
  body:
    parent: { database_id: "{{ .database_id }}" }
    properties:
      Name:
        title:
          - { text: { content: "{{ .title }}" } }
input_schema:
  type: object
  required: [database_id, title]
  properties:
    database_id: { type: string }
    title: { type: string }
output_schema:
  type: object
permissions:
  network:
    - host: api.notion.com
      port: 443
  timeout_seconds: 15
```

### Validate it

```bash
cog-gear-lint my_notion_gear.yaml
# OK: declaration matches cog_gear: v1 schema
```

### Submit it

Submit a PR to your cog deployment's gear catalogue (each deployment
manages its own — there is no central "cog catalogue"). The cog
host registers the gear on the next restart or admin-UI reload.

---

## The principle

The validator in this repo enforces **exactly** the same envelope
the closed-source cog engine enforces at load time. If `cog-gear-lint`
says your gear is valid, cog will load it. If they ever drift,
that's a bug — file an issue.

The Tier envelope rules:

- **Tier 1 (`tier: http`)**: required `network`; `timeout_seconds`
  capped at engine maximum (default 60s); `subprocess: true` rejected;
  `file_read` or `file_write` non-empty rejected.
- **Tier 2 (`tier: webhook`)**: same envelope as Tier 1, plus required
  `external_workflow:` block declaring the workflow platform.

Operators may further restrict (per-user `allowed_gear_tiers` in cog
policy). The validator's job is the lower bound: "would this even
load?" Whether the operator has *granted* the tier to the user is a
runtime question this repo doesn't answer.

---

## What's NOT here

- The cog engine source. Closed-source proprietary — see
  [TIERS.md](https://cog.ai/docs/tiers) and the cog distribution.
- The Tier 0 gear catalogue. Native gears (bash, fetch, file ops,
  git, build/test, office_read/write, etc.) ship inside the engine.
- Runtime policy enforcement. Even if your gear is schema-valid,
  whether a given user may invoke it depends on the operator's
  policy and the user's licence tier.

---

## Contributing

PRs welcome for:
- Example gears (one PR per service).
- `cog-gear-lint` bug fixes.
- Documentation improvements.

For changes to the schema itself — the meta-schema is versioned
(`cog_gear: v1`). Breaking changes ship as `v2`; the cog engine
maintains parser support for both during a transition window.

---

## License

MIT — see [`LICENSE`](./LICENSE).

The MIT licence applies to:
- The schema files (`cog_gear.v1.schema.{yaml,json}`).
- The `cog-gear-lint` validator.
- The example gear declarations.
- The authoring documentation.

It does **not** extend to the cog engine itself, which is governed
by the proprietary licence at
[github.com/GreyAssoc/cog/LICENSE](https://github.com/GreyAssoc/cog/blob/main/LICENSE).

---

**Maintained by:** Grey & Associates Ltd
**Contact:** gears@cog.ai
