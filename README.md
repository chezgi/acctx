# acctx

`acctx` is a Linux-first CLI for creating and maintaining an AI-agent-friendly accounting and compliance workspace for Iranian companies.

## Product boundary

P0 provides project initialization, embedded content, managed agent instructions, skill distribution, project inspection, deterministic filesystem plans, skill overrides, and explicit content upgrades.

It is **not** an accounting ledger, ERP, ETL engine, AI runtime, or government-submission client. External tools such as Codex or Claude Code perform AI-based interpretation through the materialized skills. Later phases add task workspaces, deterministic calculators, exports, and Iran-specific compliance skills.

## Requirements

- Linux
- An existing Git repository
- Git remains the rollback/history mechanism; `acctx` does not create commits or modify branches/remotes.

## Build

```bash
make verify
./bin/acctx version
```

## Initialize an existing Git repository

```bash
cd /path/to/company-repo
acctx init --plan
acctx init --non-interactive --yes
```

`init` creates `.acctx/manifest.yaml`, company bootstrap files, vendor content, and relative skill symlinks.

## Agent integration

- `AGENTS.md` is the canonical shared instruction file.
- `CLAUDE.md` imports it with `@AGENTS.md`.
- Claude skills: `.claude/skills/acctx-*`
- Codex skills: `.agents/skills/acctx-*`
- Canonical vendor skills: `skills/vendor/`
- Company overrides: `skills/company/`

Existing text outside `<!-- acctx:begin -->` / `<!-- acctx:end -->` is preserved.

## Project inspection

```bash
acctx project status
acctx project doctor
acctx content version
acctx content list
```

## Skill lifecycle

```bash
acctx skill list
acctx skill status acctx-workspace
acctx skill validate acctx-workspace
acctx skill override acctx-workspace --plan
acctx skill override acctx-workspace --yes
acctx skill diff acctx-workspace
acctx skill reset acctx-workspace --force --yes
```

## Upgrade

```bash
acctx project upgrade --plan
acctx project upgrade --yes
```

P0 only performs a no-op when the embedded content version matches the pinned project version; cross-version migration is intentionally conservative until migration rules are shipped.

## Git and security policy

All company data is intended to be Git-trackable except secrets. The managed `.gitignore` block excludes only:

```text
.acctx/cache/
.acctx/staging/
.acctx/tmp/
```

Never commit passwords, API tokens, private keys, signing keys, portal sessions, or unencrypted private certificate keys.

## P1 and P2 are not implemented in P0

P0 does not implement fiscal-year commands, task workspaces, generic imports, accounting ledgers, tax calculators, official exports, or government submission. These are separate phases and should remain small, task-oriented extensions.

## Research-source policy

The research article supplied during design is retained only as non-authoritative research input. It is not an approved legal, accounting, or architectural authority for this project.
