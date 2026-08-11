# ACCTX P0.5 Core Content Implementation Plan

> **For agentic workers:** implement task-by-task with tests before production changes.

**Goal:** Make `acctx init` produce a useful Iran-focused accounting skill workspace rather than only two bootstrap skills.

**Architecture:** Keep the CLI AI-free. Add a versioned embedded content catalog, materialize vendor skills/workflows/templates/references, and preserve the existing manifest, symlink, drift, and override model.

**Tech Stack:** Go 1.23 standard library, embedded files, Linux symlinks, GitHub Actions.

## Global Constraints

- No AI runtime or external model API in `acctx`.
- No accounting ledger, ERP, or generic ETL subsystem.
- All standard skills use the `acctx-` prefix.
- The uploaded research article remains `non_authoritative` and `not_approved`.
- Existing company-owned files are never overwritten by vendor materialization.

## Tasks

1. Expand the embedded catalog and the `ir-software-kb-techpark` preset to the core skill set.
2. Add domain skills with purpose, required inputs, AI workflow, deterministic-tool boundary, outputs, and final controls.
3. Add shared workflows, task templates, source governance, and the research-source classification.
4. Materialize `workflows/vendor`, `templates/vendor`, and `references/vendor` during init/upgrade planning.
5. Add integration tests proving vendor assets and provider skill links are created.
6. Run `make verify`, open a PR, and merge only after CI succeeds.
