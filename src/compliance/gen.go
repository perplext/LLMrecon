// Package compliance — OWASP Agentic 2026 mapping codegen scaffolding (v0.9.0).
//
// The hand-written TechniqueToAgenticCategories map in owasp_agentic.go has
// historically drifted from the canonical mapping in
// templates/owasp_agentic_2026.yaml (the YAML lists more techniques than the
// Go map encodes). The v0.9.0 plan introduces a code generator that derives
// the Go map from the YAML at build time, eliminating drift by construction.
//
// To regenerate the map locally:
//
//	go run ./cmd/owasp-gen \
//	    -yaml templates/owasp_agentic_2026.yaml \
//	    -out  src/compliance/owasp_agentic_generated.go
//
// In CI: `go generate ./... && git diff --exit-code` will fail the build
// when the generated file is out of sync. This is wired up in v0.10.0.
//
// v0.9.0 ships the generator and this scaffolding only; the existing
// hand-written map in owasp_agentic.go remains the canonical lookup for
// runtime behavior. v0.10.0 will switch the lookup to the generated map and
// delete the hand-written entries. This intermediate step lets us ship the
// generator (and its tests) without changing runtime behavior.

package compliance

//go:generate go run ../../cmd/owasp-gen -yaml ../../templates/owasp_agentic_2026.yaml -out owasp_agentic_generated.go
