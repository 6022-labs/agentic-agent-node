package settings

import (
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const AgentIdentitySettingsKey = "agent_identity"

// DefaultEnsBaseDomain is the parent zone used when ens_base_domain is unset.
const DefaultEnsBaseDomain = "6022.eth"

const defaultIdentitySyncIntervalSeconds = 600

// AgentIdentitySettings holds the agent's public URL (published as the ENS `url`
// record) and the base zone its display name is built under.
type AgentIdentitySettings struct {
	RuntimeUrl           string
	EnsBaseDomain        string
	IdentitySyncInterval time.Duration
}

type agentIdentitySettingsRaw struct {
	RuntimeUrl                  string `koanf:"runtime_url"`
	EnsBaseDomain               string `koanf:"ens_base_domain"`
	IdentitySyncIntervalSeconds int    `koanf:"identity_sync_interval_seconds"`
}

func NewAgentIdentitySettings(logger *zap.Logger, k *koanf.Koanf) *AgentIdentitySettings {
	raw := agentIdentitySettingsRaw{IdentitySyncIntervalSeconds: defaultIdentitySyncIntervalSeconds}
	if err := k.Unmarshal(AgentIdentitySettingsKey, &raw); err != nil {
		logger.Fatal("failed to unmarshal agent identity settings", zap.Error(err))
	}

	raw.RuntimeUrl = strings.TrimSpace(raw.RuntimeUrl)
	if len(raw.RuntimeUrl) == 0 {
		logger.Fatal("please set agent_identity.runtime_url (or AGENT_IDENTITY__RUNTIME_URL env)")
	}

	raw.EnsBaseDomain = strings.TrimSpace(raw.EnsBaseDomain)
	if len(raw.EnsBaseDomain) == 0 {
		raw.EnsBaseDomain = DefaultEnsBaseDomain
	}

	if raw.IdentitySyncIntervalSeconds <= 0 {
		logger.Warn("invalid agent_identity.identity_sync_interval_seconds value, defaulting to 600")
		raw.IdentitySyncIntervalSeconds = defaultIdentitySyncIntervalSeconds
	}

	return &AgentIdentitySettings{
		RuntimeUrl:           raw.RuntimeUrl,
		EnsBaseDomain:        raw.EnsBaseDomain,
		IdentitySyncInterval: time.Duration(raw.IdentitySyncIntervalSeconds) * time.Second,
	}
}
