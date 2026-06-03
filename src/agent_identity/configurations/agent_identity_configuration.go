package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"go.uber.org/dig"
)

func AddAgentIdentityConfiguration(container *dig.Container) {
	if err := container.Provide(services.NewBootstrapRunner); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewNftOwnerAuthorizer,
		dig.As(new(common_interfaces.AuthorizerInterface)),
	); err != nil {
		panic(err)
	}

	useCaseProviders := []any{
		use_cases.NewAuthorize,
		use_cases.NewGetAgentIdentity,
		use_cases.NewPublishEnsRuntimeRecords,
	}
	for _, p := range useCaseProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}
}
