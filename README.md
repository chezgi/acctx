# acctx

`acctx` is a Linux-first CLI for creating and maintaining an AI-agent-friendly accounting and compliance workspace for Iranian companies.

## Product boundary

`acctx` provides project initialization, embedded Iran-focused skills, managed agent instructions, company/year/task scaffolding, deterministic validation and reconciliation calculators, evidence indexing, controlled draft bundles, skill overrides, and explicit content upgrades.

It is **not** an accounting ledger, ERP, generic ETL engine, AI runtime, legal decision-maker, or government-submission client. Codex, Claude Code, or another external agent interprets irregular company files through materialized skills. `acctx` owns only deterministic workspace, calculation, validation, indexing, and export operations.

## Requirements

- Linux
- An existing Git repository
- Git remains the rollback/history mechanism; `acctx` does not create commits or modify branches/remotes.

## Build and verify

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

`init` creates:

- `.acctx/manifest.yaml`;
- editable company bootstrap files;
- 28 vendor skills for the `ir-software-kb-techpark` preset;
- vendor workflows, templates, references, and annual rules;
- relative Claude Code and Codex skill symlinks.

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

A fiscal year is a context folder. Creating it does not create a ledger, opening entries, or balance roll-forward.

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

A task workspace connects original inputs, one domain skill, optional templates, deterministic calculation results, AI-generated drafts, and a final checklist. Original company files remain free-form.

## Standard CSV validation

Calculators accept only standard task CSV formats materialized by `task init`.

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

Only purchase rows classified as `eligible-credit` are included as input credit. A valid calculation does not prove legal eligibility.

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

Only adjustment rows marked `approved` or `accepted` are included. Benefits and credits require evidence and human review before being supplied.

### Payroll-tax reconciliation

```bash
acctx calc payroll-tax \
  --year 1405 \
  --input accounting/fiscal-years/1405/work/payroll-tax/templates/input.csv \
  --output accounting/fiscal-years/1405/work/payroll-tax/calculations/result.json \
  --json
```

This calculator covers ordinary progressive annual salary income. Special-category, flat-rate, and separately exempt items must be classified outside the calculator.

### Statutory-baseline deadlines

```bash
acctx calc deadline-events
acctx calc deadline \
  --event tax-assessment-objection \
  --date 1405-05-10 \
  --json
```

Deadline output is a statutory baseline. It does not silently apply holidays, working-day rules, temporary extensions, or special notification facts.

## Evidence indexes

```bash
acctx evidence index \
  --year 1405 \
  --task vat-q1 \
  --output accounting/fiscal-years/1405/work/vat-q1/calculations/evidence-index.json \
  --json
```

Optional scope flags:

```text
--include-company
--include-year-inputs
```

The index records project-relative paths, categories, media types, sizes, and SHA-256 values. Symlinks and non-regular files are rejected.

## Controlled draft bundles

List formats:

```bash
acctx export formats
```

Generic task bundle:

```bash
acctx export task \
  --year 1405 \
  --task vat-q1 \
  --output accounting/fiscal-years/1405/outputs/vat-q1.zip \
  --include-company \
  --json
```

Audit pack wrapper:

```bash
acctx export audit-pack \
  --year 1404 \
  --task audit \
  --output accounting/fiscal-years/1404/outputs/audit-pack.zip \
  --json
```

Tax pack wrapper:

```bash
acctx export tax-pack \
  --year 1399 \
  --task tax-defense \
  --output accounting/fiscal-years/1399/outputs/tax-defense.zip \
  --json
```

Existing ZIP files are protected unless `--force` is supplied.

Verify a bundle independently:

```bash
acctx export verify \
  --input accounting/fiscal-years/1399/outputs/tax-defense.zip \
  --json
```

Every bundle contains:

```text
bundle-manifest.json
evidence-index.json
files/<project-relative-path>
```

Every bundle is marked `draft`, `submission_performed: false`, and `final_human_review_required: true`. Technical verification does not approve or submit the package.

## Annual rule content

Embedded rules cover 1397 through 1405 and are materialized at:

```text
rules/vendor/ir/annual.json
references/vendor/rule-sources.yaml
```

Algorithms are type-safe Go code; annual rates, brackets, scope notes, and source identifiers are versioned data. Filing-time verification against official sources remains mandatory.

## Project inspection and upgrade

```bash
acctx project status
acctx project doctor
acctx content version
acctx content list

acctx project upgrade --plan
acctx project upgrade --non-interactive --yes
```

Upgrade is explicit and drift-aware. Unmodified vendor content may be updated; modified vendor files become conflicts. Company-owned files and full skill overrides remain preserved.

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

## Git and security policy

All company data is intended to be Git-trackable except secrets. The managed `.gitignore` block excludes only:

```text
.acctx/cache/
.acctx/staging/
.acctx/tmp/
```

Never commit passwords, API tokens, private keys, signing keys, portal sessions, or unencrypted private certificate keys. Export only the smallest necessary bundle scope.

## Research-source policy

The research article supplied during design is retained only as non-authoritative research input. It is not an approved legal, accounting, or architectural authority for this project.
