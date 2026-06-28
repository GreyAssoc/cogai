# Authoring a Tier 1 / Tier 2 Cog Gear

> A practical guide to writing declarative gears that load on cog at
> runtime. Pairs with the canonical meta-schema at
> [`cog_gear.v1.schema.yaml`](../cog_gear.v1.schema.yaml).

---

## 1. What kind of gear do you want?

| You want to... | Tier | Notes |
|---|---|---|
| Call a single REST API that needs auth | **1 (http)** | The most common case. Notion, Stripe, Linear, GitHub, internal services. |
| Fire a webhook into Zapier / n8n / Make / IFTTT | **2 (webhook)** | Same wire shape, plus admission. Operator must opt into webhooks. |
| Run shell commands, touch files, do anything not HTTP | **Tier 0** | You need a Go gear in the engine source. Open a feature request. |

This guide covers tiers 1 and 2.

---

## 2. The minimum-viable Tier 1 gear

Save as `my_gear.yaml`:

```yaml
cog_gear: v1
tier: http
name: echo_status
description: Echo the HTTP status code from httpbin.
endpoint:
  method: GET
  url: "https://httpbin.org/status/{{ .code }}"
input_schema:
  type: object
  required: [code]
  properties:
    code: { type: integer, minimum: 100, maximum: 599 }
output_schema:
  type: object
permissions:
  network:
    - host: httpbin.org
      port: 443
  timeout_seconds: 5
```

Validate it:

```bash
cog-gear-lint my_gear.yaml
```

That's the minimum. Everything else is refinement.

---

## 3. Required fields, explained

### `cog_gear: v1`

The declaration format version. Always `v1` today. When v2 ships,
both will work for a transition window.

### `tier: http` (or `webhook`)

Determines which envelope rules apply. `http` is direct REST;
`webhook` adds the `external_workflow:` admission.

### `name`

Lowercase snake_case. 3-64 chars. Must be unique within the
deployment. This is how the model invokes the gear (`@gear name`).

**Naming convention:** `<service>_<verb>_<noun>` reads naturally to
the model.

- `notion_create_page`
- `stripe_get_customer`
- `github_list_issues`
- `linear_close_issue`

Bad names: `do_thing`, `helper`, `notion` (too generic).

### `description`

Read by the model when deciding to invoke. Be specific.

**Good:** "Create a new row in a Notion database. Provide
`database_id` and `properties` (a map of property name → value)."

**Bad:** "Notion gear."

Aim for 30-300 characters. Include input shape hints so the model
doesn't have to read the input schema to know whether the gear
applies.

### `endpoint`

The HTTP request shape:

- `method`: GET / POST / PUT / PATCH / DELETE.
- `url`: full HTTPS URL. Path params via `{{ .name }}` templating.
  Use the templating engine's path-safe escaping for IDs.
- `query`: query parameters. Static or templated.
- `headers`: request headers. Use `${secret:NAME}` for credentials.
- `body`: request body. Templated; JSON-encoded by default.

### `input_schema`

JSON Schema for the gear's input. The model sees this as the tool's
input shape. Use `properties`, `required`, `type`, `enum`, `pattern`,
`minimum`, `maximum`, `minLength`, `maxLength` — the standard
vocabulary.

Examples that read well to a model:

```yaml
input_schema:
  type: object
  required: [database_id, title]
  properties:
    database_id:
      type: string
      description: "Notion database ID (32 hex chars without dashes)."
      pattern: "^[a-f0-9]{32}$"
    title:
      type: string
      description: "Page title."
      minLength: 1
      maxLength: 200
    tags:
      type: array
      items: { type: string }
      description: "Optional tags to add to the new page."
```

The `description` on each property is read by the model. Use it to
explain semantics, not just shape.

### `output_schema`

Optional but recommended. Tells the model what the result will
look like.

```yaml
output_schema:
  type: object
  properties:
    page_id: { type: string }
    url: { type: string }
```

If the response shape is unknown, `{ "type": "object" }` is fine.

### `permissions`

The required permissions for Tier 1:

```yaml
permissions:
  network:
    - host: api.notion.com
      port: 443
  timeout_seconds: 15
```

- `network`: array of `host`/`port` pairs the gear may call. The
  loader compares the URL's host against this list; mismatches
  refused at load time.
- `timeout_seconds`: per-call timeout. Engine maximum default is
  60s; declarations over the maximum are rejected.
- `max_calls_per_minute`: optional per-user rate limit.

**Forbidden in Tier 1 and Tier 2:**

- `subprocess: true`
- non-empty `file_read`
- non-empty `file_write`

The loader rejects these regardless of what you declare.

---

## 4. Templating

Two templating sources in a gear declaration:

### `{{ .input_field }}` — input substitution

Render at call time from the model's input. Equivalent to Go's
`text/template` with the input as the root value.

```yaml
url: "https://api.notion.com/v1/pages/{{ .page_id }}"
body:
  properties:
    Name:
      title:
        - { text: { content: "{{ .title }}" } }
```

### `${secret:NAME}` — secret substitution

Resolved at load time from the operator's configured secret source
(env vars, Vault, AWS Secrets Manager, etc.). Never serialised into
traces.

```yaml
headers:
  Authorization: "Bearer ${secret:NOTION_TOKEN}"
```

**Why two syntaxes?** Input substitution happens at call time and
varies per call. Secret substitution happens at load time and is
constant. Mixing them by accident is a security bug; the two
distinct syntaxes prevent it.

---

## 5. Tier 2 (webhook) specifics

Webhook gears add the `external_workflow:` block:

```yaml
cog_gear: v1
tier: webhook
name: trigger_zap_new_customer
description: Notify the new-customer Zap with email + plan.
endpoint:
  method: POST
  url: "https://hooks.zapier.com/hooks/catch/${secret:ZAP_NEW_CUSTOMER_ID}/"
  body:
    email: "{{ .email }}"
    plan: "{{ .plan }}"
input_schema:
  type: object
  required: [email, plan]
  properties:
    email: { type: string, format: email }
    plan: { type: string, enum: [free, pro, teams] }
permissions:
  network:
    - host: hooks.zapier.com
      port: 443
  timeout_seconds: 10
external_workflow:
  platform: zapier
  workflow_id: "zap_new_customer_v3"
```

Notes:

- The operator must enable `AllowExternalWorkflowGears` for the
  loader to accept webhook gears.
- The model-facing description gets a "via external workflow:
  zapier" prefix automatically.
- The admin UI shows a badge.
- The trace records `external_workflow: zapier` for audit.

Use Tier 2 when the actual API call goes to a workflow platform. Do
NOT use it to relabel a Tier 1 gear; the loader's envelope check is
the same either way, but Tier 2 carries visible degradation that
should mean what it says.

---

## 6. Common patterns

### Bearer auth

```yaml
headers:
  Authorization: "Bearer ${secret:SERVICE_TOKEN}"
```

### Basic auth

```yaml
headers:
  Authorization: "Basic ${secret:SERVICE_BASIC_AUTH_B64}"
```

(`B64` suffix is a hint — the operator pre-encodes the credential.)

### API key in query

```yaml
query:
  api_key: "${secret:SERVICE_KEY}"
```

### JSON request body

```yaml
headers:
  Content-Type: application/json
body:
  field1: "{{ .input_field1 }}"
  field2: "{{ .input_field2 }}"
```

### Form-encoded request body

```yaml
headers:
  Content-Type: application/x-www-form-urlencoded
body: "field1={{ .input_field1 }}&field2={{ .input_field2 }}"
```

### Pagination via cursor

Two approaches:

- **One call returns one page** — the model can call the gear
  repeatedly with successive cursors.
- **Gear handles pagination internally** — only feasible for Tier 0;
  Tier 1 maps one call to one HTTP request.

For Tier 1, prefer the first. Model agents handle iterative calls.

---

## 7. Anti-patterns

- **One gear that does everything for a service.** Split. Notion is
  ten gears (create_page, update_page, search, etc.), not one.
- **Vague descriptions.** "Notion gear" doesn't help the model.
- **Embedding secrets directly.** Always `${secret:NAME}`.
- **Path-templating the host.** The host is immutable; only the
  path and query may template.
- **Timeouts beyond the engine maximum.** Will be rejected at load.
- **Tier 1 declaring `subprocess: true`.** Rejected.
- **Tier 2 without `external_workflow:`.** Rejected.

---

## 8. Submitting a gear

Each cog deployment manages its own gear catalogue. Submit a PR to
your operator's gear repo or admin UI (varies by deployment); do
not submit to this public repo — it holds the published schema and
reference examples, not a central catalogue.

Issues against the schema, the `cog-gear-lint` validator binary,
or this authoring guide are welcome via
[github.com/GreyAssoc/cogai/issues](https://github.com/GreyAssoc/cogai/issues).

---

## 9. Validating

Run `cog-gear-lint` before submitting:

```bash
cog-gear-lint my_notion_gear.yaml
```

Validates:

- Required fields present and non-empty.
- Tier envelope rules (Tier 1 / 2 forbid subprocess, file_read,
  file_write).
- Tier 2 requires `external_workflow`; Tier 1 forbids it.
- Name format (snake_case, 3-64 chars).
- URL host matches `permissions.network[].host`.

Validates exactly what the cog engine's loader checks. A passing
declaration loads cleanly; if it doesn't, that's a bug — file an
issue.

---

**Last updated:** 2026-06-05