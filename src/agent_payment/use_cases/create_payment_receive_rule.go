package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

type CreatePaymentReceiveRule struct {
	*AbstractPaymentUseCase
}

func NewCreatePaymentReceiveRule(
	agentRepository identity_repositories.AgentRepositoryInterface,
	agentAddressRepository identity_repositories.AgentAddressRepositoryInterface,
	paymentReceivePolicyRepository repositories.PaymentReceivePolicyRepositoryInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface,
	token6022Settings *runtime_common_settings.Token6022Settings,
) *CreatePaymentReceiveRule {
	return &CreatePaymentReceiveRule{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			agentRepository,
			agentAddressRepository,
			paymentReceivePolicyRepository,
			paymentReceiveRuleRepository,
			paymentReceiveMethodStrategies,
			token6022Settings,
		),
	}
}

func (u *CreatePaymentReceiveRule) Execute(
	ctx context.Context,
	request *requests.SavePaymentReceiveRuleRequest,
) (*responses.PaymentReceiveRuleResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	strategy, ok := u.paymentReceiveMethodStrategies.For(enums.PaymentMethod(request.Method))
	if !ok {
		return nil, errors.NewValidationError("method", "unsupported payment method")
	}
	if err := strategy.Validate(ctx, request); err != nil {
		return nil, err
	}

	policy, _, err := u.ensurePaymentReceivePolicy(ctx)
	if err != nil {
		return nil, err
	}

	rule, err := u.paymentReceiveRuleRepository.Create(ctx, u.toPaymentReceiveRule(policy, request))
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err.Error())
	}
	// Persist the method config after the rule (it needs the rule id); roll the
	// rule back if that fails so we never leave an orphan without config.
	if err := strategy.SaveConfig(ctx, rule.Id, request.Config); err != nil {
		_ = u.paymentReceiveRuleRepository.DeleteById(ctx, rule.Id)
		return nil, errors.NewInternalError("create_failed", err.Error())
	}
	return responses.ToPaymentReceiveRuleResponse(rule, request.Config), nil
}

func (u *CreatePaymentReceiveRule) toPaymentReceiveRule(
	policy *models.PaymentReceivePolicy,
	request *requests.SavePaymentReceiveRuleRequest,
) *models.PaymentReceiveRule {
	return &models.PaymentReceiveRule{
		PolicyId: policy.Id,
		Name:     request.Name,
		Price:    request.Price,
		Method:   enums.PaymentMethod(request.Method),
	}
}
