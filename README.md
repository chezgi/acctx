# acctx

`acctx` is a Linux-first CLI for creating and maintaining an AI-agent-friendly accounting and compliance workspace for Iranian companies.

## Product boundary

`acctx` provides project initialization, embedded Iran-focused skills, managed agent instructions, project inspection, lightweight company/year/task scaffolding, deterministic filesystem plans, skill overrides, and explicit content upgrades.

It is **not** an accounting ledger, ERP, generic ETL engine, AI runtime, or government-submission client. Codex, Claude Code, or another external agent interprets irregular company files through the materialized skills. `acctx` only owns deterministic workspace operations and, in later phases, narrow validation/calculation/export commands.

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

`init` creates `.acctx/manifest.yaml`, editable company bootstrap files, 27 vendor skills for the `ir-software-kb-techpark` preset, vendor workflows/templates/references, and relative Claude Code/Codex skill symlinks.

## Agent integration

- `AGENTS.md` is the canonical shared instruction file.
- `CLAUDE.md` imports it with `@AGENTS.md`.
- Claude skills: `.claude/skills/acctx-*`
- Codex skills: `.agents/skills/acctx-*`
- Canonical vendor skills: `skills/vendor/`
- Company overrides: `skills/company/`

Existing text outside `<!-- acctx:begin -->` / `<!-- acctx:end -->` is preserved.

## Company profile

Company YAML files under `accounting/company/` are company-owned after creation and may be edited normally.

```bash
acctx company status
acctx company validate --profile bootstrap
acctx company validate --profile complete
acctx company validate --profile tax-ready
acctx company validate --profile vat-ready
```

## Fiscal years

A fiscal year is only a folder/context descriptor. Creating it does not create a ledger, opening entries, or balance roll-forward.

```bash
acctx year init 1405 --non-interactive --yes
acctx year init 1399 --historical --non-interactive --yes
acctx year init 1396 --archive --non-interactive --yes
acctx year init 1401-transition \
  --mode reconstruction \
  --start 1401-07-01 \
  --end 1401-12-29 \
  --ruleset-year 1401 \
  --non-interactive --yes

acctx year list
acctx year status 1405
```

Each year contains free-form `inputs/`, task `work/`, and final `outputs/` directories. Years before 1397 are archive-only.

## Task workspaces

```bash
acctx task types
acctx task init vat --year 1405 --period Q1 --non-interactive --yes
acctx task init corporate-tax --year 1404 --non-interactive --yes
acctx task init tax-defense --year 1399 --non-interactive --yes
acctx task list --year 1405
acctx task status vat-q1 --year 1405
```

A task workspace connects original inputs, one domain skill, optional standard templates, deterministic calculation results, AI-generated drafts, and a final checklist. Original company files remain free-form.

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
acctx skill status acctx-vat
acctx skill validate acctx-vat
acctx skill override acctx-vat --plan
acctx skill override acctx-vat --yes
acctx skill diff acctx-vat
acctx skill reset acctx-vat --force --yes
```

## Upgrade

```bash
acctx project upgrade --plan
acctx project upgrade --yes
```

Cross-version updates remain conservative: company-owned files and company skill overrides are never silently overwritten.

## Git and security policy

All company data is intended to be Git-trackable except secrets. The managed `.gitignore` block excludes only:

```text
.acctx/cache/
.acctx/staging/
.acctx/tmp/
```

Never commit passwords, API tokens, private keys, signing keys, portal sessions, or unencrypted private certificate keys.

## Remaining deterministic phases

P2 adds narrow validators and calculators for standard task templates. P3 adds controlled draft exporters and evidence bundles. Neither phase will add an internal AI runtime or direct government submission.

## Research-source policy

The research article supplied during design is retained only as non-authoritative research input. It is not an approved legal, accounting, or architectural authority for this project.
