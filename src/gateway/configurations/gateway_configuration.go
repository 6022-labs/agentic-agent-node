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

	if err := container.Provide(use_cases.NewGetAuthChallenge); err != nil {
		panic(err)
	}
	if err := container.Provide(use_cases.NewVerifyAuth); err != nil {
		panic(err)
	}
}

func newNonceStore(jwtSettings *settings.JwtSettings) *services.NonceStore {
	return services.NewNonceStore(jwtSettings.NonceTtl)
}
