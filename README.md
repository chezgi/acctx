# acctx

`acctx` is a Linux-first CLI for creating and maintaining an AI-agent-friendly accounting and compliance workspace for Iranian companies.

## Product boundary

`acctx` provides project initialization, embedded Iran-focused skills, managed agent instructions, project inspection, lightweight company/year/task scaffolding, deterministic filesystem plans, standard-input validation, narrow tax reconciliations, skill overrides, and explicit content upgrades.

It is **not** an accounting ledger, ERP, generic ETL engine, AI runtime, or government-submission client. Codex, Claude Code, or another external agent interprets irregular company files through the materialized skills. `acctx` only owns deterministic workspace operations and narrow validation/calculation/export commands.

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

`init` creates `.acctx/manifest.yaml`, editable company bootstrap files, 27 vendor skills for the `ir-software-kb-techpark` preset, vendor workflows/templates/references/rules, and relative Claude Code/Codex skill symlinks.

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

## Standard CSV validation

Calculators accept only the standard task CSV formats materialized by `task init`.

```bash
acctx validate kinds
acctx validate vat \
  --input accounting/fiscal-years/1405/work/vat-q1/templates/input.csv

acctx validate corporate-tax \
  --input accounting/fiscal-years/1404/work/corporate-tax/templates/adjustments.csv

acctx validate payroll-tax \
  --input accounting/fiscal-years/1405/work/payroll-tax/templates/input.csv
```

All monetary values are integer IRR. Invalid or unresolved classifications return exit code `2`.

## Deterministic calculators

### VAT reconciliation

```bash
acctx calc vat \
  --year 1405 \
  --input accounting/fiscal-years/1405/work/vat-q1/templates/input.csv \
  --output accounting/fiscal-years/1405/work/vat-q1/calculations/result.json \
  --json
```

The calculator reconciles output VAT and only includes purchase rows classified as `eligible-credit`. It reports general-rate mismatches but does not decide legal eligibility for the agent or user.

### Corporate income-tax reconciliation

```bash
acctx calc corporate-tax \
  --year 1404 \
  --book-profit 10000000000 \
  --tax-credits 500000000 \
  --input accounting/fiscal-years/1404/work/corporate-tax/templates/adjustments.csv \
  --output accounting/fiscal-years/1404/work/corporate-tax/calculations/result.json \
  --json
```

Only adjustment rows marked `approved` or `accepted` are included. Exemptions, credits, knowledge-based benefits, technology-park benefits, and R&D credits must be supported and reviewed before they are supplied to the calculator.

### Payroll-tax reconciliation

```bash
acctx calc payroll-tax \
  --year 1405 \
  --input accounting/fiscal-years/1405/work/payroll-tax/templates/input.csv \
  --output accounting/fiscal-years/1405/work/payroll-tax/calculations/result.json \
  --json
```

This calculator covers ordinary progressive annual salary income. Special-category, flat-rate, and separately exempt items must be classified outside the calculator before use.

### Statutory-baseline deadlines

```bash
acctx calc deadline-events
acctx calc deadline \
  --event tax-assessment-objection \
  --date 1405-05-10 \
  --json
```

Deadline output is a **statutory baseline**. It does not silently apply holidays, working-day rules, temporary extensions, or special service/notification facts. Final verification is required.

## Annual rule content

Embedded rules cover 1397 through 1405 and are materialized at:

```text
rules/vendor/ir/annual.json
references/vendor/rule-sources.yaml
```

Algorithms are type-safe Go code; annual rates, brackets, scope notes, and source identifiers are versioned data. Installing a new binary does not silently rewrite a company project. Filing-time verification against official sources remains mandatory.

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

## Remaining deterministic phase

P3 adds controlled draft exporters, task/evidence manifests, and portable audit/tax bundles. It will not add an internal AI runtime, direct government submission, or a generic reporting database.

## Research-source policy

The research article supplied during design is retained only as non-authoritative research input. It is not an approved legal, accounting, or architectural authority for this project.
