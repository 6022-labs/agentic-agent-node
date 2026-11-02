package use_cases

import (
	"context"
	"encoding/json"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/google/uuid"
)

// AbstractPaymentUseCase carries what every receive use case needs; materialization lives in the provider.
type AbstractPaymentUseCase struct {
	paymentReceivePolicyProvider      payment_interfaces.PaymentReceivePolicyProviderInterface
	paymentReceiveRuleRepository      repositories.PaymentReceiveRuleRepositoryInterface
	paymentReceiveRulePriceRepository repositories.PaymentReceiveRulePriceRepositoryInterface
	paymentReceiveMethodStrategies    payment_interfaces.PaymentReceiveMethodStrategiesInterface
	unitOfWorkFactory                 runtime_common_interfaces.UnitOfWorkFactoryInterface
}

func NewAbstractPaymentUseCase(
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveRulePriceRepository repositories.PaymentReceiveRulePriceRepositoryInterface,
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *AbstractPaymentUseCase {
	return &AbstractPaymentUseCase{
		paymentReceivePolicyProvider:      paymentReceivePolicyProvider,
		paymentReceiveRuleRepository:      paymentReceiveRuleRepository,
		paymentReceiveRulePriceRepository: paymentReceiveRulePriceRepository,
		paymentReceiveMethodStrategies:    paymentReceiveMethodStrategies,
		unitOfWorkFactory:                 unitOfWorkFactory,
	}
}

func (u *AbstractPaymentUseCase) ensurePaymentReceivePolicy(
	ctx context.Context,
) (*models.PaymentReceivePolicy, *identity_models.Agent, error) {
	return u.paymentReceivePolicyProvider.Ensure(ctx)
}

// claimAsset fails when another rule already prices the token; ruleId may already own it.
func (u *AbstractPaymentUseCase) claimAsset(
	ctx context.Context,
	strategy payment_interfaces.PaymentReceiveMethodStrategyInterface,
	policyId uuid.UUID,
	ruleId uuid.UUID,
	config json.RawMessage,
) error {
	ownerId, err := strategy.RuleIdForAsset(ctx, policyId, config)
	if err != nil {
		return errors.NewInternalError("rule_lookup_failed", err.Error())
	}
	if ownerId != uuid.Nil && ownerId != ruleId {
		return errors.NewConflictError("asset_already_has_rule", "another rule already prices this asset")
	}
	return nil
}

func (u *AbstractPaymentUseCase) createPrices(
	ctx context.Context,
	ruleId uuid.UUID,
	requested []requests.SavePaymentReceiveRulePriceRequest,
) ([]models.PaymentReceiveRulePrice, error) {
	prices := make([]models.PaymentReceiveRulePrice, 0, len(requested))
	for i := range requested {
		price, err := u.paymentReceiveRulePriceRepository.Create(ctx, &models.PaymentReceiveRulePrice{
			PaymentReceiveRuleId: ruleId,
			Price:                requested[i].Price,
			AddressPatterns:      requested[i].AddressPatterns,
		})
		if err != nil {
			return nil, err
		}
		prices = append(prices, *price)
	}
	return prices, nil
}
