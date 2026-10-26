package use_cases

import (
	"context"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/google/uuid"
)

// AbstractPaymentUseCase carries what every receive use case needs; materialization lives in the provider.
type AbstractPaymentUseCase struct {
	paymentReceivePolicyProvider   payment_interfaces.PaymentReceivePolicyProviderInterface
	paymentReceiveRuleRepository   repositories.PaymentReceiveRuleRepositoryInterface
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface
}

func NewAbstractPaymentUseCase(
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface,
) *AbstractPaymentUseCase {
	return &AbstractPaymentUseCase{
		paymentReceivePolicyProvider:   paymentReceivePolicyProvider,
		paymentReceiveRuleRepository:   paymentReceiveRuleRepository,
		paymentReceiveMethodStrategies: paymentReceiveMethodStrategies,
	}
}

func (u *AbstractPaymentUseCase) ensurePaymentReceivePolicy(
	ctx context.Context,
) (*models.PaymentReceivePolicy, *identity_models.Agent, error) {
	return u.paymentReceivePolicyProvider.Ensure(ctx)
}

func (u *AbstractPaymentUseCase) evmAddress(ctx context.Context, agentId uuid.UUID) (string, error) {
	return u.paymentReceivePolicyProvider.EvmAddress(ctx, agentId)
}
