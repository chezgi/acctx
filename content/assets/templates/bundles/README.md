# Controlled bundles

An `acctx` bundle contains:

- `bundle-manifest.json`: draft-only control metadata;
- `evidence-index.json`: project-relative files, categories, sizes, media types, and SHA-256 values;
- `files/`: selected source and work-product files using their project-relative paths.

The bundle is not a submission receipt, digital signature, approval, or proof that legal rules are current. Verify it with `acctx export verify` before delivery.
