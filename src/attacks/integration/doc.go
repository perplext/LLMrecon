// Package integration runs cross-family smoke tests for v0.9.0 attack
// modules. Tests gate on RUN_INTEGRATION=1 via t.Skip() so the default
// `go test ./...` invocation stays silent — operators opt in
// explicitly.
//
// The blank imports below trigger each attack package's init() blocks,
// which register module instances with attacks.DefaultRegistry. Without
// these imports the registry would be empty and DefaultRegistry.Get()
// would return "not registered" errors.
package integration

import (
	_ "github.com/perplext/LLMrecon/src/attacks/adaptive"
	_ "github.com/perplext/LLMrecon/src/attacks/memory"
	_ "github.com/perplext/LLMrecon/src/attacks/multimodal"
	_ "github.com/perplext/LLMrecon/src/attacks/reasoning"
)
