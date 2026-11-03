package use_cases

import (
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/google/uuid"
)

// AbstractPaymentSpendUseCase carries what every spend rule use case needs; materialization lives in the provider.
type AbstractPaymentSpendUseCase struct {
	paymentSpendPolicyProvider       payment_interfaces.PaymentSpendPolicyProviderInterface
	paymentSpendRuleRepository       repositories.PaymentSpendRuleRepositoryInterface
	paymentSpendRuleWindowRepository repositories.PaymentSpendRuleWindowRepositoryInterface
	paymentSpendMethodStrategies     payment_interfaces.PaymentSpendMethodStrategiesInterface
	unitOfWorkFactory                runtime_common_interfaces.UnitOfWorkFactoryInterface
}

func NewAbstractPaymentSpendUseCase(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendRuleWindowRepository repositories.PaymentSpendRuleWindowRepositoryInterface,
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *AbstractPaymentSpendUseCase {
	return &AbstractPaymentSpendUseCase{
		paymentSpendPolicyProvider:       paymentSpendPolicyProvider,
		paymentSpendRuleRepository:       paymentSpendRuleRepository,
		paymentSpendRuleWindowRepository: paymentSpendRuleWindowRepository,
		paymentSpendMethodStrategies:     paymentSpendMethodStrategies,
		unitOfWorkFactory:                unitOfWorkFactory,
	}
}

// claimAsset fails when another rule already spends the token; owned reports that ruleId spends it already.
func (u *AbstractPaymentSpendUseCase) claimAsset(
	ctx context.Context,
	paymentSpendMethodStrategy payment_interfaces.PaymentSpendMethodStrategyInterface,
	policyId uuid.UUID,
	ruleId uuid.UUID,
	config json.RawMessage,
) (owned bool, err error) {
	ownerId, err := paymentSpendMethodStrategy.RuleIdForAsset(ctx, policyId, config)
	if err != nil {
		return false, errors.NewInternalError("rule_lookup_failed", err.Error())
	}
	if ownerId != uuid.Nil && ownerId != ruleId {
		return false, errors.NewConflictError("asset_already_has_rule", "another rule already limits this asset")
	}
	return ownerId != uuid.Nil, nil
}
