package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"go.uber.org/dig"
)

const paymentReceiveMethodStrategiesGroup = "payment_receive_method_strategies"

const paymentSpendMethodStrategiesGroup = "payment_spend_method_strategies"

func AddAgentPaymentConfiguration(container *dig.Container) {
	useCases := []any{
		use_cases.NewGetPaymentReceivePolicy,
		use_cases.NewDeletePaymentReceivePolicy,
		use_cases.NewGetPaymentFacilitators,
		use_cases.NewListPaymentReceiveRules,
		use_cases.NewCreatePaymentReceiveRule,
		use_cases.NewUpdatePaymentReceiveRule,
		use_cases.NewDeletePaymentReceiveRule,
		use_cases.NewSignX402Payment,
		use_cases.NewAuthorizePaymentReceiveAccess,
		use_cases.NewGetPaymentSpendPolicy,
		use_cases.NewListPaymentSpendRules,
		use_cases.NewCreatePaymentSpendRule,
		use_cases.NewUpdatePaymentSpendRule,
		use_cases.NewDeletePaymentSpendRule,
		use_cases.NewDeletePaymentSpendPolicy,
	}
	for _, useCase := range useCases {
		if err := container.Provide(useCase); err != nil {
			panic(err)
		}
	}

	if err := container.Provide(settings.NewPaymentSpendSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewX402PaymentSigner,
		dig.As(new(payment_interfaces.X402PaymentSignerInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewPaymentReceiveMethodStrategies,
		dig.As(new(payment_interfaces.PaymentReceiveMethodStrategiesInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewX402PaymentReceiveMethodStrategy,
		dig.Group(paymentReceiveMethodStrategiesGroup),
		dig.As(new(payment_interfaces.PaymentReceiveMethodStrategyInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewPaymentReceivePolicyProvider,
		dig.As(new(payment_interfaces.PaymentReceivePolicyProviderInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		services.NewX402PaymentReceivePolicyEvaluator,
		dig.As(new(payment_interfaces.X402PaymentReceivePolicyEvaluatorInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		services.NewX402PaymentVerifier,
		dig.As(new(payment_interfaces.X402PaymentVerifierInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewPaymentSpendMethodStrategies,
		dig.As(new(payment_interfaces.PaymentSpendMethodStrategiesInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewX402PaymentSpendMethodStrategy,
		dig.Group(paymentSpendMethodStrategiesGroup),
		dig.As(new(payment_interfaces.PaymentSpendMethodStrategyInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewPaymentSpendPolicyProvider,
		dig.As(new(payment_interfaces.PaymentSpendPolicyProviderInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewX402PaymentSpendPolicyEvaluator,
		dig.As(new(payment_interfaces.X402PaymentSpendPolicyEvaluatorInterface)),
	); err != nil {
		panic(err)
	}
}
