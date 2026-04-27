package configurations

import "go.uber.org/dig"

// AddAgentReasoningConfiguration is a no-op today — the package is scaffolded
// ahead of P2.4 (LLM exec, MCP exec, inner loop). Wire registrations land here
// when those services are implemented.
func AddAgentReasoningConfiguration(container *dig.Container) {
	_ = container
}
