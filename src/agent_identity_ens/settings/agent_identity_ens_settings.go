package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const AgentIdentityEnsSettingsKey = "agent_identity_ens"

// AgentIdentityEnsSettings holds operator-supplied values published to ENS at
// boot. RuntimeUrl is the public URL this agent is reachable at; it is written
// to the `agent.runtime-url` text record on the agent's ENS subnode.
type AgentIdentityEnsSettings struct {
	RuntimeUrl string `koanf:"runtime_url"`
}

func NewAgentIdentityEnsSettings(logger *zap.Logger, k *koanf.Koanf) *AgentIdentityEnsSettings {
	var s AgentIdentityEnsSettings
	if err := k.Unmarshal(AgentIdentityEnsSettingsKey, &s); err != nil {
		logger.Fatal("failed to unmarshal agent_identity_ens settings", zap.Error(err))
	}

	s.RuntimeUrl = strings.TrimSpace(s.RuntimeUrl)
	if len(s.RuntimeUrl) == 0 {
		logger.Fatal("please set agent_identity_ens.runtime_url (or AGENT_IDENTITY_ENS__RUNTIME_URL env)")
	}

	return &s
}
