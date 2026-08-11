# ACCTX P2 Deterministic Calculators Implementation Plan

> **For agentic workers:** implement each task test-first and merge only after GitHub Actions verifies the complete branch.

**Goal:** Add narrow, versioned, deterministic validation and calculation commands for standard task files from 1397 through 1405.

**Architecture:** Annual rule parameters are embedded JSON and materialized for inspection. Go implements typed algorithms for progressive payroll tax reconciliation, VAT reconciliation, corporate-tax reconciliation, and Jalali statutory-baseline deadlines. Irregular source interpretation remains the responsibility of external AI skills.

**Tech Stack:** Go 1.23 standard library, CSV/JSON, embedded files, integer IRR arithmetic, GitHub Actions.

## Global Constraints

- Never use floating point for money or tax rates.
- Calculators operate only on standard task CSV files.
- Payroll calculation covers ordinary progressive salary income; special-category and flat-rate items require separate human classification.
- Deadline output is a statutory baseline and does not silently apply holidays or temporary extensions.
- Rule source metadata must state that filing-time verification is required.
- No AI runtime, ledger, generic importer, or direct government submission.

## Tasks

1. Add annual rules for 1397–1405 and materialize `rules/vendor`.
2. Add integer basis-point and progressive bracket algorithms with golden tests.
3. Add standard CSV schema validation.
4. Add VAT reconciliation calculator.
5. Add corporate income-tax reconciliation calculator.
6. Add annual payroll-tax reconciliation calculator.
7. Add Jalali date arithmetic and statutory-baseline deadline calculator.
8. Add `validate` and `calc` CLI commands and optional JSON output files.
9. Add integration tests, documentation, CI verification, PR, and merge.
