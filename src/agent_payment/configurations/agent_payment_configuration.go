package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"go.uber.org/dig"
)

func AddAgentPaymentConfiguration(container *dig.Container) {
	providers := []any{
		use_cases.NewGetPaymentPolicy,
		use_cases.NewDeletePaymentPolicy,
		use_cases.NewGetPaymentFacilitators,
		use_cases.NewListPaymentRules,
		use_cases.NewCreatePaymentRule,
		use_cases.NewUpdatePaymentRule,
		use_cases.NewDeletePaymentRule,
	}
	for _, p := range providers {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}
}
