package services

import (
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
)

// RuntimeRecordsProvider assembles the agent's ENS text records from the
// contract-fixed keys and the runtime config (the public URL).
type RuntimeRecordsProvider struct {
	agentIdentitySettings *settings.AgentIdentitySettings
}

func NewRuntimeRecordsProvider(agentIdentitySettings *settings.AgentIdentitySettings) *RuntimeRecordsProvider {
	return &RuntimeRecordsProvider{agentIdentitySettings: agentIdentitySettings}
}

func (p *RuntimeRecordsProvider) RuntimeRecords() []identity_interfaces.EnsRecord {
	return []identity_interfaces.EnsRecord{
		{Key: identity_interfaces.UrlRecordKey, Value: p.agentIdentitySettings.RuntimeUrl},
	}
}
