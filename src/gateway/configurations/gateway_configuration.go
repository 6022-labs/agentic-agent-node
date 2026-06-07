package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/gateway/services"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/settings"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"go.uber.org/dig"
)

func AddGatewayConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewJwtSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		newNonceStore,
		dig.As(new(gateway_interfaces.NonceStoreInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewSignatureVerifier,
		dig.As(new(gateway_interfaces.SignatureVerifierInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewJwtIssuer,
		dig.As(new(gateway_interfaces.JwtIssuerInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewJwtValidator,
		dig.As(new(gateway_interfaces.JwtValidatorInterface)),
	); err != nil {
		panic(err)
	}

	useCaseProviders := []any{
		use_cases.NewGetAuthChallenge,
		use_cases.NewVerifyAuth,
		use_cases.NewGetLlmConfig,
		use_cases.NewSetLlmConfig,
		use_cases.NewDeleteLlmConfig,
		use_cases.NewListMcpConnections,
		use_cases.NewCreateMcpConnection,
		use_cases.NewDeleteMcpConnection,
		use_cases.NewSaveAgentRegistration,
		use_cases.NewGetAgentRegistration,
		use_cases.NewGetAgentMintInfo,
		use_cases.NewCheckAgentMint,
		use_cases.NewGetAgentEnsInfo,
		use_cases.NewCheckAgentEns,
		use_cases.NewSelfRegisterAgent,
		use_cases.NewListConversations,
		use_cases.NewListConversationMessages,
	}
	for _, p := range useCaseProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}

	if err := container.Provide(
		use_cases.NewCreateResponse,
		dig.As(new(use_cases.CreateResponseInterface)),
	); err != nil {
		panic(err)
	}
}

func newNonceStore(jwtSettings *settings.JwtSettings) *services.NonceStore {
	return services.NewNonceStore(jwtSettings.NonceTtl)
}
