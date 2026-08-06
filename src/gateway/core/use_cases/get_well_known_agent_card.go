package use_cases

import (
	"context"
	"strings"

	core_settings "github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/responses"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	identity_settings "github.com/6022-labs/agentic-agent-node/src/gateway/identity/settings"
)

const (
	a2aProtocolVersion = "1.0"
	a2aProtocolBinding = "JSONRPC"
	a2aAgentVersion    = "1.0.0"
	a2aChatSkillId     = "chat"
	a2aMessageRoute    = "/a2a"
)

type GetWellKnownAgentCard struct {
	wellKnownCardSigner

	agentRegistrationClient identity_interfaces.AgentRegistrationClientInterface
	agentIdentitySettings   *identity_settings.AgentIdentitySettings
	gatewaySettings         *core_settings.GatewaySettings
}

func NewGetWellKnownAgentCard(
	agentRegistrationClient identity_interfaces.AgentRegistrationClientInterface,
	agentCardSignatureClient identity_interfaces.AgentCardSignatureClientInterface,
	agentIdentitySettings *identity_settings.AgentIdentitySettings,
	gatewaySettings *core_settings.GatewaySettings,
) *GetWellKnownAgentCard {
	return &GetWellKnownAgentCard{
		wellKnownCardSigner:     wellKnownCardSigner{agentCardSignatureClient: agentCardSignatureClient},
		agentRegistrationClient: agentRegistrationClient,
		agentIdentitySettings:   agentIdentitySettings,
		gatewaySettings:         gatewaySettings,
	}
}

// Execute fetches the agent identity from the runtime; the card requires it, so a failure surfaces rather than a placeholder.
func (u *GetWellKnownAgentCard) Execute(ctx context.Context) (*responses.WellKnownAgentCard, error) {
	payload, err := u.agentRegistrationClient.Get(ctx)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(payload.Name)

	card := &responses.WellKnownAgentCard{
		Name:        name,
		Description: name + " — 6022 agent reachable over A2A",
		SupportedInterfaces: []responses.AgentInterface{{
			Url:             u.agentIdentitySettings.RuntimeUrl + u.gatewaySettings.PublicPrefix + a2aMessageRoute,
			ProtocolBinding: a2aProtocolBinding,
			ProtocolVersion: a2aProtocolVersion,
		}},
		Version: a2aAgentVersion,
		Capabilities: responses.AgentCapabilities{
			Streaming:         false,
			PushNotifications: false,
		},
		DefaultInputModes:  []string{"text/plain"},
		DefaultOutputModes: []string{"text/plain"},
		Skills: []responses.AgentSkill{{
			Id:          a2aChatSkillId,
			Name:        name,
			Description: "Conversational inference turn",
			Tags:        []string{"chat"},
			InputModes:  []string{"text/plain"},
			OutputModes: []string{"text/plain"},
		}},
	}

	signatures, err := u.signCard(ctx, card)
	if err != nil {
		return nil, err
	}
	card.Signatures = signatures

	return card, nil
}
