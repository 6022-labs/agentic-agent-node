package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"go.uber.org/dig"
)

const paymentMethodStrategiesGroup = "payment_method_strategies"

func AddAgentPaymentConfiguration(container *dig.Container) {
	useCases := []any{
		use_cases.NewGetPaymentReceivePolicy,
		use_cases.NewDeletePaymentReceivePolicy,
		use_cases.NewGetPaymentFacilitators,
		use_cases.NewListPaymentReceiveRules,
		use_cases.NewCreatePaymentReceiveRule,
		use_cases.NewUpdatePaymentReceiveRule,
		use_cases.NewDeletePaymentReceiveRule,
	}
	for _, useCase := range useCases {
		if err := container.Provide(useCase); err != nil {
			panic(err)
		}
	}

	if err := container.Provide(
		services.NewPaymentMethodStrategies,
		dig.As(new(payment_interfaces.PaymentMethodStrategiesInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		services.NewX402PaymentMethodStrategy,
		dig.Group(paymentMethodStrategiesGroup),
		dig.As(new(payment_interfaces.PaymentMethodStrategyInterface)),
	); err != nil {
		panic(err)
	}
}
