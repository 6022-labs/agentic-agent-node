package configurations

import "go.uber.org/dig"

// AddAgentReasoningHttpInferenceConfiguration is a no-op today — the package is
// scaffolded ahead of P2.4 (HTTP client → LLM providers). Wire registrations
// land here when the client is implemented.
func AddAgentReasoningHttpInferenceConfiguration(container *dig.Container) {
	_ = container
}
