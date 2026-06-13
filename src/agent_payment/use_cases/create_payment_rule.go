package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

type CreatePaymentRule struct {
	*AbstractPaymentUseCase
	erc20Validator payment_interfaces.Erc20ValidatorInterface
}

func NewCreatePaymentRule(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
	blockchainSettings *runtime_common_settings.BlockchainSettings,
	token6022Settings *runtime_common_settings.Token6022Settings,
	paymentFacilitatorsSettings *runtime_common_settings.PaymentFacilitatorsSettings,
	erc20Validator payment_interfaces.Erc20ValidatorInterface,
) *CreatePaymentRule {
	return &CreatePaymentRule{
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

func (u *CreatePaymentRule) Execute(ctx context.Context, request *requests.SavePaymentRuleRequest) (*responses.PaymentRuleResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	if err := u.requireFacilitatorForPaidRule(request); err != nil {
		return nil, err
	}

	policy, _, err := u.ensurePaymentPolicy(ctx)
	if err != nil {
		return nil, err
	}
	if err := u.validateAsset(ctx, request); err != nil {
		return nil, err
	}

	rule, err := u.paymentRuleRepository.Create(ctx, toPaymentRule(policy, request))
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err.Error())
	}
	return responses.ToPaymentRuleResponse(rule), nil
}

func (u *CreatePaymentRule) validateAsset(ctx context.Context, request *requests.SavePaymentRuleRequest) error {
	chainId, err := requests.ParseEip155ChainId(request.Network)
	if err != nil {
		return errors.NewValidationError("network", err.Error())
	}
	if err := u.erc20Validator.ValidateErc20(ctx, chainId, request.Asset); err != nil {
		return errors.NewValidationError("asset", err.Error())
	}
	return nil
}

func toPaymentRule(policy *models.PaymentPolicy, request *requests.SavePaymentRuleRequest) *models.PaymentRule {
	return &models.PaymentRule{
		PolicyId:          policy.Id,
		Name:              request.Name,
		Network:           request.Network,
		Asset:             request.Asset,
		PayTo:             request.PayTo,
		Price:             request.Price,
		MaxTimeoutSeconds: request.MaxTimeoutSeconds,
		Eip712Name:        request.Eip712Name,
		Eip712Version:     request.Eip712Version,
		AddressPatterns:   request.AddressPatterns,
	}
}
