# ACCTX P1 Company, Years, and Task Workspaces Implementation Plan

> **For agentic workers:** implement each task test-first and verify through GitHub Actions before merge.

**Goal:** Add lightweight company validation, fiscal-year scaffolding, and task workspaces without turning `acctx` into accounting software.

**Architecture:** Company data remains editable YAML. Fiscal years and task descriptors are JSON documents with `.yaml` names (valid YAML subsets). Mutating commands produce filesystem plans and use the existing plan/apply engine. User input files remain free-form under each year.

**Tech Stack:** Go 1.23 standard library, embedded task templates, Linux filesystem, GitHub Actions.

## Global Constraints

- No ledger, database, generic importer, or AI runtime.
- `year init` creates folders only; it does not create accounting entries or roll balances.
- `task init` connects one year, one domain skill, optional templates, and draft/output folders.
- Years before 1397 are archive-only.
- Existing user-owned year/task files are not overwritten.

## Tasks

1. Add a minimal top-level YAML scalar reader and company status/validation profiles.
2. Add fiscal-year model, safe path validation, Jalali default dates, init/list/status operations.
3. Add a task registry and init/list/status operations with embedded template copies.
4. Add `company`, `year`, and `task` CLI dispatch and plan/apply confirmation behavior.
5. Add unit and Linux integration tests.
6. Update README, run `make verify`, open a PR, and merge after successful CI.
