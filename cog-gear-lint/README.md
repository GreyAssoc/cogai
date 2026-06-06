# cog-gear-lint

A Go-based CLI that validates Tier 1 / Tier 2 cog gear declarations
against the `cog_gear: v1` schema. Run before submitting a gear to
your cog deployment's catalogue; CI integrators run it on PRs.

## Install

```bash
go install github.com/GreyAssoc/cogai/cog-gear-lint@latest
```

Or build from source:

```bash
git clone https://github.com/GreyAssoc/cogai
cd cogai/cog-gear-lint
go build .
```

## Use

```bash
cog-gear-lint my_gear.yaml
# OK: my_gear.yaml

cog-gear-lint examples/
# OK: examples/notion_create_page.yaml
# OK: examples/stripe_get_customer.yaml
# OK: examples/github_create_issue.yaml
# OK: examples/linear_list_issues.yaml
```

Exit codes:
- `0` — all declarations valid.
- `1` — one or more declarations failed validation.
- `2` — usage error (missing arg, unreadable file).

## What it checks

The validator enforces exactly the same envelope rules the cog
engine's loader applies at runtime:

- Top-level shape: `cog_gear: v1`, `tier`, `name`, `description`,
  `endpoint`, `input_schema`, `permissions` all required.
- `name`: lowercase snake_case, 3-64 chars.
- `description`: 10-1024 chars.
- `endpoint.method`: one of GET/POST/PUT/PATCH/DELETE.
- `endpoint.url`: HTTPS, host portion literal (no templates).
- `permissions.network`: non-empty array of `{host, port}`; URL host
  must appear in the list.
- `permissions.timeout_seconds`: integer 1..60.
- **Tier 1 (http)**: `subprocess`, `file_read`, `file_write`
  rejected. `external_workflow` rejected.
- **Tier 2 (webhook)**: same envelope as Tier 1, plus
  `external_workflow.platform` required.

A passing declaration loads cleanly on the cog engine. If they ever
drift, that's a bug — file an issue at the cogai repo.

## License

MIT — see [`../LICENSE`](../LICENSE).
