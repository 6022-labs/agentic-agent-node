package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const AgentIdentitySettingsKey = "agent_identity"

// DefaultEnsBaseDomain is the parent zone used when ens_base_domain is unset.
const DefaultEnsBaseDomain = "6022.eth"

// AgentIdentitySettings holds the agent's public URL (published as the ENS `url`
// record) and the base zone its display name is built under.
type AgentIdentitySettings struct {
	RuntimeUrl    string `koanf:"runtime_url"`
	EnsBaseDomain string `koanf:"ens_base_domain"`
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

	s.EnsBaseDomain = strings.TrimSpace(s.EnsBaseDomain)
	if len(s.EnsBaseDomain) == 0 {
		s.EnsBaseDomain = DefaultEnsBaseDomain
	}

	return &s
}
