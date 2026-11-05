package configurations

import (
	core_services "github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	core_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	core_settings "github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	core_use_cases "github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	identity_settings "github.com/6022-labs/agentic-agent-node/src/gateway/identity/settings"
	identity_use_cases "github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases"
	knowledge_use_cases "github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases"
	payment_use_cases "github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	reasoning_use_cases "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"go.uber.org/dig"
)

// AddGatewayConfiguration wires the gateway domain by delegating to each
// per-domain sub-configuration (core, identity, knowledge, payment, reasoning).
func AddGatewayConfiguration(container *dig.Container) {
	addCoreConfiguration(container)
	addIdentityConfiguration(container)
	addKnowledgeConfiguration(container)
	addPaymentConfiguration(container)
	addReasoningConfiguration(container)
}

func addCoreConfiguration(container *dig.Container) {
	if err := container.Provide(core_settings.NewJwtSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(core_settings.NewGatewaySettings); err != nil {
		panic(err)
	}

	if err := container.Provide(core_settings.NewSiweSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(core_settings.NewIPFSGatewaysSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		newNonceStore,
		dig.As(new(core_interfaces.NonceStoreInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		core_services.NewSignatureVerifier,
		dig.As(new(core_interfaces.SignatureVerifierInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		core_services.NewJwtIssuer,
		dig.As(new(core_interfaces.JwtIssuerInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		core_services.NewJwtValidator,
		dig.As(new(core_interfaces.JwtValidatorInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		core_services.NewRefreshTokenManager,
		dig.As(new(core_interfaces.RefreshTokenManagerInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		core_services.NewRefreshSessionStore,
		dig.As(new(core_interfaces.RefreshSessionStoreInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		core_services.NewSiweMessageBuilder,
		dig.As(new(core_interfaces.SiweMessageBuilderInterface)),
	); err != nil {
		panic(err)
	}

	useCaseProviders := []any{
		core_use_cases.NewGetAuthChallenge,
		core_use_cases.NewVerifyAuth,
		core_use_cases.NewRefreshAuth,
		core_use_cases.NewRevokeAuth,
		core_use_cases.NewGetConfig,
		core_use_cases.NewGetWellKnownAgentCard,
		core_use_cases.NewGetWellKnownAgent6022,
	}
	for _, p := range useCaseProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}
}

func addIdentityConfiguration(container *dig.Container) {
	if err := container.Provide(identity_settings.NewAgentIdentitySettings); err != nil {
		panic(err)
	}

	useCaseProviders := []any{
		identity_use_cases.NewSaveAgentRegistration,
		identity_use_cases.NewGetAgentRegistration,
		identity_use_cases.NewGetAgentMintInfo,
		identity_use_cases.NewCheckAgentMint,
		identity_use_cases.NewGetAgentEnsInfo,
		identity_use_cases.NewCheckAgentEns,
		identity_use_cases.NewSelfRegisterAgent,
		identity_use_cases.NewListAgentCollections,
		identity_use_cases.NewGetAgentSoul,
	}
	for _, p := range useCaseProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}
}

func addKnowledgeConfiguration(container *dig.Container) {
	useCaseProviders := []any{
		knowledge_use_cases.NewListConversations,
		knowledge_use_cases.NewCreateConversation,
		knowledge_use_cases.NewUpdateConversation,
		knowledge_use_cases.NewListConversationMessages,
		knowledge_use_cases.NewAppendConversationMessages,
	}
	for _, p := range useCaseProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}
}

func addPaymentConfiguration(container *dig.Container) {
	useCaseProviders := []any{
		payment_use_cases.NewGetPaymentReceivePolicy,
		payment_use_cases.NewDeletePaymentReceivePolicy,
		payment_use_cases.NewListPaymentReceiveRules,
		payment_use_cases.NewCreatePaymentReceiveRule,
		payment_use_cases.NewUpdatePaymentReceiveRule,
		payment_use_cases.NewDeletePaymentReceiveRule,
		payment_use_cases.NewGetPaymentSpendPolicy,
		payment_use_cases.NewUpdatePaymentSpendPolicy,
		payment_use_cases.NewDeletePaymentSpendPolicy,
		payment_use_cases.NewListPaymentSpendRules,
		payment_use_cases.NewCreatePaymentSpendRule,
		payment_use_cases.NewUpdatePaymentSpendRule,
		payment_use_cases.NewDeletePaymentSpendRule,
	}
	for _, p := range useCaseProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}
}

func addReasoningConfiguration(container *dig.Container) {
	useCaseProviders := []any{
		reasoning_use_cases.NewGetLlmConfig,
		reasoning_use_cases.NewSetLlmConfig,
		reasoning_use_cases.NewDeleteLlmConfig,
		reasoning_use_cases.NewValidateLlmConfig,
		reasoning_use_cases.NewGetAgentGuidance,
		reasoning_use_cases.NewSetAgentGuidance,
		reasoning_use_cases.NewListMcpConnections,
		reasoning_use_cases.NewDiscoverMcpServerTools,
		reasoning_use_cases.NewCreateMcpConnection,
		reasoning_use_cases.NewUpdateMcpConnection,
		reasoning_use_cases.NewDeleteMcpConnection,
		reasoning_use_cases.NewListSelfTools,
		reasoning_use_cases.NewUpdateSelfTool,
	}
	for _, p := range useCaseProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}

	if err := container.Provide(
		reasoning_use_cases.NewCreateResponse,
		dig.As(new(reasoning_use_cases.CreateResponseInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		reasoning_use_cases.NewSendA2AMessage,
		dig.As(new(reasoning_use_cases.SendA2AMessageInterface)),
	); err != nil {
		panic(err)
	}
}

func newNonceStore(jwtSettings *core_settings.JwtSettings) *core_services.NonceStore {
	return core_services.NewNonceStore(jwtSettings.NonceTtl)
}
