package configurations

import "go.uber.org/dig"

// AddAgentReasoningHttpMcpConfiguration is a no-op today — the package is
// scaffolded ahead of P2.4 (HTTP client → MCP servers). Wire registrations
// land here when the client is implemented.
func AddAgentReasoningHttpMcpConfiguration(container *dig.Container) {
	_ = container
}
