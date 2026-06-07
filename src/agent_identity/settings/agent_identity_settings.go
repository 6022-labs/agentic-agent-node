package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const AgentIdentitySettingsKey = "agent_identity"

// AgentIdentitySettings holds the public URL this agent is reachable at. The
// domain publishes it as the `url` ENS text record on the agent's subnode.
type AgentIdentitySettings struct {
	RuntimeUrl string `koanf:"runtime_url"`
}

func NewAgentIdentitySettings(logger *zap.Logger, k *koanf.Koanf) *AgentIdentitySettings {
	var s AgentIdentitySettings
	if err := k.Unmarshal(AgentIdentitySettingsKey, &s); err != nil {
		logger.Fatal("failed to unmarshal agent identity settings", zap.Error(err))
	}

	s.RuntimeUrl = strings.TrimSpace(s.RuntimeUrl)
	if len(s.RuntimeUrl) == 0 {
		logger.Fatal("please set agent_identity.runtime_url (or AGENT_IDENTITY__RUNTIME_URL env)")
	}

	return &s
}
