package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
)

// AbstractPaymentSpendUseCase carries what every spend use case needs; materialization lives in the provider.
type AbstractPaymentSpendUseCase struct {
	paymentSpendPolicyProvider   payment_interfaces.PaymentSpendPolicyProviderInterface
	paymentSpendRuleRepository   repositories.PaymentSpendRuleRepositoryInterface
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface
}

func NewAbstractPaymentSpendUseCase(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface,
) *AbstractPaymentSpendUseCase {
	return &AbstractPaymentSpendUseCase{
		paymentSpendPolicyProvider:   paymentSpendPolicyProvider,
		paymentSpendRuleRepository:   paymentSpendRuleRepository,
		paymentSpendMethodStrategies: paymentSpendMethodStrategies,
	}
}
