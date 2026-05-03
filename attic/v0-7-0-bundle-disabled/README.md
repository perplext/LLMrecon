# Atticked v0.7.0-era bundle command stubs

These five `bundle_*.go.disabled` files were written against an early
draft of the `src/bundle` package API and have been out of sync with
the live API for many releases:

- `bundle_import.go.disabled` — references `bundle.LoadBundleManifest`,
  `bundle.ImportEvent`, `bundle.ImportEventStart`, etc., none of which
  exist in the current `src/bundle` package.
- `bundle_verify.go.disabled` — references signature-verification APIs
  whose shapes have since changed; the live verification path is in
  `src/bundle/signature.go::VerifyBundle`.
- `bundle_publish.go.disabled`, `bundle_sync.go.disabled`,
  `bundle_registry.go.disabled` — three commands that depend on a
  bundle-registry concept the codebase never finished. Aspirational.

## v0.10.0 #177 disposition

`bundle import` and `bundle verify` were rewritten fresh against the
current API and live in `src/cmd/bundle_import.go` and
`src/cmd/bundle_verify.go` respectively. The originals are kept here
as historical reference for anyone restoring the publish/sync/registry
flows in v0.11.0+.

If you're picking these up: don't try to lift the `.disabled` files
verbatim. They were written when `BundleImporter` had a
`SetProgressHandler` method and `ImportOptions` had `TemplatesDir` /
`ModulesDir` / `PreserveStructure` / `Verbose` fields, none of which
the current API surfaces. Treat these files as a sketch of intent
and write a fresh implementation against the live `src/bundle` API.
