package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/google/uuid"
)

type UpdatePaymentRule struct {
	*AbstractPaymentUseCase
	erc20Validator payment_interfaces.Erc20ValidatorInterface
}

func NewUpdatePaymentRule(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
	blockchainSettings *runtime_common_settings.BlockchainSettings,
	token6022Settings *runtime_common_settings.Token6022Settings,
	paymentFacilitatorsSettings *runtime_common_settings.PaymentFacilitatorsSettings,
	erc20Validator payment_interfaces.Erc20ValidatorInterface,
) *UpdatePaymentRule {
	return &UpdatePaymentRule{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			agentRepository,
			paymentPolicyRepository,
			paymentRuleRepository,
			blockchainSettings,
			token6022Settings,
			paymentFacilitatorsSettings,
		),
		erc20Validator: erc20Validator,
	}
}

// Execute replaces every field of the rule; identity, position and policy
// binding are preserved.
func (u *UpdatePaymentRule) Execute(ctx context.Context, id uuid.UUID, request *requests.SavePaymentRuleRequest) (*responses.PaymentRuleResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	if err := u.requireFacilitatorForPaidRule(request); err != nil {
		return nil, err
	}

	rule, err := u.paymentRuleRepository.GetById(ctx, id)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if rule == nil {
		return nil, errors.NewNotFoundError("rule_not_found", "")
	}

	chainId, err := requests.ParseEip155ChainId(request.Network)
	if err != nil {
		return nil, errors.NewValidationError("network", err.Error())
	}
	if err := u.erc20Validator.ValidateErc20(ctx, chainId, request.Asset); err != nil {
		return nil, errors.NewValidationError("asset", err.Error())
	}

	rule.Name = request.Name
	rule.Network = request.Network
	rule.Asset = request.Asset
	rule.PayTo = request.PayTo
	rule.Price = request.Price
	rule.MaxTimeoutSeconds = request.MaxTimeoutSeconds
	rule.Eip712Name = request.Eip712Name
	rule.Eip712Version = request.Eip712Version
	rule.AddressPatterns = request.AddressPatterns

	updated, err := u.paymentRuleRepository.Update(ctx, rule)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	return responses.ToPaymentRuleResponse(updated), nil
}
