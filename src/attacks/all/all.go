// Package all is a registration barrel for every attack-module package
// in src/attacks/. Importing this package triggers each module package's
// init() block, which registers AttackModule instances with
// attacks.DefaultRegistry.
//
// Without this barrel, `attacks.DefaultRegistry` is empty at runtime in
// any binary that doesn't directly import each attack package. Until
// v0.10.0 issue #173, only the integration smoke-test package imported
// attack packages, so the production CLI binary saw an empty registry.
//
// Adding a new attack-module package: append a blank import below in
// alphabetical order, run `go test ./src/cmd/... -run TestNoNameCollisions`
// to verify no duplicate-name registrations, and update CLAUDE.md if the
// new category warrants documentation.
//
// Packages without registrations (`exfiltration`, `extraction`,
// `injection`, `jailbreak`, `payloads`, `persistence`, `supply_chain`)
// are deliberately omitted — the post-v0.9.0 review (issue F7) flagged
// these as legacy import-orphans. Adding them here would inflate binary
// size without populating the registry. Triage of each lives in v0.11.0
// scope; until then, omitting is correct.
package all

import (
	_ "github.com/perplext/LLMrecon/src/attacks/adaptive"
	_ "github.com/perplext/LLMrecon/src/attacks/agentic/browser"
	_ "github.com/perplext/LLMrecon/src/attacks/agentic/deception"
	_ "github.com/perplext/LLMrecon/src/attacks/agentic/mcp"
	_ "github.com/perplext/LLMrecon/src/attacks/agentic/multi_agent"
	_ "github.com/perplext/LLMrecon/src/attacks/agentic/persistence"
	_ "github.com/perplext/LLMrecon/src/attacks/agentic/skill_injection"
	_ "github.com/perplext/LLMrecon/src/attacks/agentic/tool_use"
	_ "github.com/perplext/LLMrecon/src/attacks/audio"
	_ "github.com/perplext/LLMrecon/src/attacks/evasion"
	_ "github.com/perplext/LLMrecon/src/attacks/memory"
	_ "github.com/perplext/LLMrecon/src/attacks/multimodal"
	_ "github.com/perplext/LLMrecon/src/attacks/orchestration"
	_ "github.com/perplext/LLMrecon/src/attacks/rag"
	_ "github.com/perplext/LLMrecon/src/attacks/reasoning"
)
