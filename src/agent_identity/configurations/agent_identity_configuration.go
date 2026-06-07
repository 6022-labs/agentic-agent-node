package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"go.uber.org/dig"
)

func AddAgentIdentityConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewAgentIdentitySettings); err != nil {
		panic(err)
	}

	if err := container.Provide(services.NewBootstrapRunner); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewRuntimeRecordsProvider,
		dig.As(new(identity_interfaces.RuntimeRecordsProviderInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewMintPublisher,
		dig.As(new(identity_interfaces.MintPublisherInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewEnsPublisher,
		dig.As(new(identity_interfaces.EnsPublisherInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewNftOwnerAuthorizer,
		dig.As(new(common_interfaces.AuthorizerInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewRegistrationReconciler,
		dig.As(new(identity_interfaces.RegistrationReconcilerInterface)),
	); err != nil {
		panic(err)
	}

	useCaseProviders := []any{
		use_cases.NewAuthorize,
		use_cases.NewGetAgentIdentity,
		use_cases.NewPublishEnsRuntimeRecords,
		use_cases.NewSaveAgentRegistration,
		use_cases.NewGetAgentRegistration,
		use_cases.NewGetAgentMintInfo,
		use_cases.NewCheckAgentMint,
		use_cases.NewGetAgentEnsInfo,
		use_cases.NewCheckAgentEns,
		use_cases.NewSelfRegisterAgent,
	}
	for _, p := range useCaseProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}
}
