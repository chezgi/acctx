# acctx controlled bundle format 1.0.0

The ZIP entry set is traversal-safe, sorted, and uses stable entry timestamps. Source files are indexed and re-hashed while the archive is written. The manifest must declare `status: draft`, `submission_performed: false`, `final_human_review_required: true`, and `rule_verification_required: true`.

Verification checks duplicate entries, unsafe paths, symlinks, manifest/index digests, indexed file presence, byte sizes, and SHA-256 values. A valid technical verification does not establish legal sufficiency or authorize submission.
