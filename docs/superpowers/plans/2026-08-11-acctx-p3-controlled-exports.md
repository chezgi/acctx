# ACCTX P3 Controlled Exports Implementation Plan

> **For agentic workers:** implement each task test-first and merge only after GitHub Actions verifies the complete branch.

**Goal:** Produce portable, verifiable draft task/audit/tax bundles and make explicit embedded-content upgrades usable without adding submission automation.

**Architecture:** Evidence indexing walks only selected project-owned roots, rejects symlinks and special files, and records project-relative SHA-256 metadata. Exporters create deterministic-entry ZIP files containing the selected task/context files, an evidence index, and a draft-only bundle manifest. Verification re-hashes every archive member. Project upgrade reuses the existing drift-aware workspace planner.

**Tech Stack:** Go 1.23 standard library, archive/zip, JSON, SHA-256, Linux filesystem, GitHub Actions.

## Global Constraints

- Every bundle is `draft`; `submission_performed` is always false.
- Export does not call government portals, sign documents, or approve output.
- Symlinks and non-regular files are rejected from bundle scope.
- Archive entry names must be project-relative and traversal-safe.
- Company profile and year-level inputs are opt-in for generic task bundles and enabled by default for audit/tax pack wrappers.
- Existing bundle files are not overwritten unless `--force` is supplied.
- Company-owned files and skill overrides are never silently overwritten during upgrade.
- No AI runtime, ledger, database, or generic ETL subsystem.

## Tasks

1. Add a sorted evidence index with file hashes, categories, media types, and a stable source digest.
2. Add task, audit-pack, and tax-pack ZIP writers with draft metadata and atomic output replacement.
3. Add bundle verification with duplicate, traversal, missing, size, and hash checks.
4. Add `evidence index` and `export` CLI command families.
5. Replace conflict-only cross-version upgrade with drift-aware explicit materialization.
6. Add the evidence-bundle skill, controlled-export workflow, content version 0.4.0, tests, and documentation.
7. Run `make verify`, open a PR, and merge only after successful CI.
