package services

import (
	"fmt"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
)

// EnsDomainBuilder composes the display domain matching the ENS gateway convention
// and the registry's node derivation, using the configured base zone.
type EnsDomainBuilder struct {
	agentIdentitySettings *settings.AgentIdentitySettings
}

func NewEnsDomainBuilder(agentIdentitySettings *settings.AgentIdentitySettings) *EnsDomainBuilder {
	return &EnsDomainBuilder{
		agentIdentitySettings: agentIdentitySettings,
	}
}

func (b *EnsDomainBuilder) Build(name, collectionName, chainId string) string {
	return fmt.Sprintf("%s.%s.%s.%s",
		strings.ToLower(name),
		strings.ToLower(collectionName),
		chainId,
		strings.ToLower(b.agentIdentitySettings.EnsBaseDomain),
	)
}
