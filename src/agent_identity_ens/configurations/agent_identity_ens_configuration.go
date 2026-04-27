package configurations

import "go.uber.org/dig"

// AddAgentIdentityEnsConfiguration is a no-op today — the package is scaffolded
// ahead of P2.12 (boot-time ENS text-record writer for `agent.runtime-url` +
// `agent.runtime-kind`). Wire registrations land here when the writer is
// implemented.
func AddAgentIdentityEnsConfiguration(container *dig.Container) {
	_ = container
}
