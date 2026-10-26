package use_cases

import (
	"context"
	"errors"
	"slices"
	"strings"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	core_settings "github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/responses"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	identity_settings "github.com/6022-labs/agentic-agent-node/src/gateway/identity/settings"
	payment_models "github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/samber/lo"
)

const (
	protocol6022        = "6022"
	protocol6022Version = "0.1"
	responsesRoute      = "/responses"
)

type GetWellKnownAgent6022 struct {
	wellKnownCardSigner

	agentRegistrationClient  identity_interfaces.AgentRegistrationClientInterface
	paymentReceiveRuleClient payment_interfaces.PaymentReceiveRuleClientInterface
	agentIdentitySettings    *identity_settings.AgentIdentitySettings
	gatewaySettings          *core_settings.GatewaySettings
}

func NewGetWellKnownAgent6022(
	agentRegistrationClient identity_interfaces.AgentRegistrationClientInterface,
	agentCardSignatureClient identity_interfaces.AgentCardSignatureClientInterface,
	paymentReceiveRuleClient payment_interfaces.PaymentReceiveRuleClientInterface,
	agentIdentitySettings *identity_settings.AgentIdentitySettings,
	gatewaySettings *core_settings.GatewaySettings,
) *GetWellKnownAgent6022 {
	return &GetWellKnownAgent6022{
		wellKnownCardSigner:      wellKnownCardSigner{agentCardSignatureClient: agentCardSignatureClient},
		agentRegistrationClient:  agentRegistrationClient,
		paymentReceiveRuleClient: paymentReceiveRuleClient,
		agentIdentitySettings:    agentIdentitySettings,
		gatewaySettings:          gatewaySettings,
	}
}

// Execute fetches the identity from the runtime; the document requires it, so a failure surfaces, not a placeholder.
func (u *GetWellKnownAgent6022) Execute(ctx context.Context) (*responses.WellKnownAgent6022, error) {
	payload, err := u.agentRegistrationClient.Get(ctx)
	if err != nil {
		return nil, err
	}

	paymentMethods, err := u.listPaymentMethods(ctx)
	if err != nil {
		return nil, err
	}

	card := &responses.WellKnownAgent6022{
		Protocol:        protocol6022,
		ProtocolVersion: protocol6022Version,
		Name:            strings.TrimSpace(payload.Name),
		EnsDomain:       payload.EnsDomain,
		Endpoints: responses.Agent6022Endpoints{
			Responses: u.agentIdentitySettings.RuntimeUrl + u.gatewaySettings.PublicPrefix + responsesRoute,
		},
		PaymentMethods: paymentMethods,
	}

	signatures, err := u.signCard(ctx, card)
	if err != nil {
		return nil, err
	}
	card.Signatures = signatures

	return card, nil
}

// listPaymentMethods derives distinct rule methods; a missing or not-yet-minted policy means no payment, not an error.
func (u *GetWellKnownAgent6022) listPaymentMethods(ctx context.Context) ([]string, error) {
	rules, err := u.paymentReceiveRuleClient.List(ctx)
	if err != nil {
		var notFoundErr *apperrors.NotFoundError
		var conflictErr *apperrors.ConflictError
		if errors.As(err, &notFoundErr) || errors.As(err, &conflictErr) {
			return nil, nil
		}
		return nil, err
	}

	methods := lo.Uniq(lo.Map(rules, func(rule payment_models.ExternalPaymentReceiveRule, _ int) string {
		return rule.Method
	}))

	slices.Sort(methods)

	return methods, nil
}
