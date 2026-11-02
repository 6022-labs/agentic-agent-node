package services

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// PaymentReceivePolicyProvider materializes the receive policy lazily; unlike spend it requires a minted identity.
type PaymentReceivePolicyProvider struct {
	agentRepository                   identity_repositories.AgentRepositoryInterface
	agentAddressRepository            identity_repositories.AgentAddressRepositoryInterface
	paymentReceivePolicyRepository    repositories.PaymentReceivePolicyRepositoryInterface
	paymentReceiveRuleRepository      repositories.PaymentReceiveRuleRepositoryInterface
	paymentReceiveRulePriceRepository repositories.PaymentReceiveRulePriceRepositoryInterface
	paymentReceiveMethodStrategies    payment_interfaces.PaymentReceiveMethodStrategiesInterface
	token6022Settings                 *runtime_common_settings.Token6022Settings
	unitOfWorkFactory                 runtime_common_interfaces.UnitOfWorkFactoryInterface
}

func NewPaymentReceivePolicyProvider(
	agentRepository identity_repositories.AgentRepositoryInterface,
	agentAddressRepository identity_repositories.AgentAddressRepositoryInterface,
	paymentReceivePolicyRepository repositories.PaymentReceivePolicyRepositoryInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveRulePriceRepository repositories.PaymentReceiveRulePriceRepositoryInterface,
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface,
	token6022Settings *runtime_common_settings.Token6022Settings,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *PaymentReceivePolicyProvider {
	return &PaymentReceivePolicyProvider{
		agentRepository:                   agentRepository,
		agentAddressRepository:            agentAddressRepository,
		paymentReceivePolicyRepository:    paymentReceivePolicyRepository,
		paymentReceiveRuleRepository:      paymentReceiveRuleRepository,
		paymentReceiveRulePriceRepository: paymentReceiveRulePriceRepository,
		paymentReceiveMethodStrategies:    paymentReceiveMethodStrategies,
		token6022Settings:                 token6022Settings,
		unitOfWorkFactory:                 unitOfWorkFactory,
	}
}

func (p *PaymentReceivePolicyProvider) Ensure(
	ctx context.Context,
) (*models.PaymentReceivePolicy, *identity_models.Agent, error) {
	agent, err := p.agentRepository.Get(ctx)
	if err != nil {
		return nil, nil, errors.NewInternalError("policy_failed", err.Error())
	}
	if agent == nil {
		return nil, nil, errors.NewInternalError("policy_failed", "agent identity is not initialized yet")
	}
	// The seed prices access to the agent's token, so there is nothing to seed before mint.
	if agent.CollectionTokenId == nil {
		return nil, nil, errors.NewConflictError("agent_not_minted", "agent identity is not minted yet")
	}

	policy, err := p.paymentReceivePolicyRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, nil, errors.NewInternalError("policy_failed", err.Error())
	}
	if policy != nil {
		return policy, agent, nil
	}

	policy, err = p.createSeededPolicy(ctx, agent)
	if err != nil {
		// A concurrent first request may have won the insert race; re-read before failing.
		existing, getErr := p.paymentReceivePolicyRepository.GetByAgentId(ctx, agent.Id)
		if getErr == nil && existing != nil {
			return existing, agent, nil
		}
		return nil, nil, err
	}
	return policy, agent, nil
}

// createSeededPolicy writes policy and seed atomically, so a failed seed is retried on the next call.
func (p *PaymentReceivePolicyProvider) createSeededPolicy(
	ctx context.Context,
	agent *identity_models.Agent,
) (*models.PaymentReceivePolicy, error) {
	evmAddress, err := p.EvmAddress(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("policy_failed", err.Error())
	}

	unitOfWork, err := p.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return nil, errors.NewInternalError("policy_failed", err.Error())
	}
	defer func() { _ = unitOfWork.Rollback() }()

	policy, err := p.paymentReceivePolicyRepository.Create(
		unitOfWork.Context(),
		&models.PaymentReceivePolicy{AgentId: agent.Id},
	)
	if err != nil {
		return nil, errors.NewInternalError("policy_failed", err.Error())
	}
	if err := p.seedDefaultRule(unitOfWork.Context(), policy, agent, evmAddress); err != nil {
		return nil, err
	}
	if err := unitOfWork.Commit(); err != nil {
		return nil, errors.NewInternalError("policy_failed", err.Error())
	}
	return policy, nil
}

func (p *PaymentReceivePolicyProvider) seedDefaultRule(
	ctx context.Context,
	policy *models.PaymentReceivePolicy,
	agent *identity_models.Agent,
	evmAddress string,
) error {
	rule, config, price, err := p.defaultRule(policy, agent, evmAddress)
	if err != nil {
		return errors.NewInternalError("policy_failed", err.Error())
	}
	if rule == nil {
		return nil
	}
	created, err := p.paymentReceiveRuleRepository.Create(ctx, rule)
	if err != nil {
		return errors.NewInternalError("policy_failed", err.Error())
	}
	if strategy, ok := p.paymentReceiveMethodStrategies.For(created.Method); ok {
		if err := strategy.SaveConfig(ctx, policy.Id, created.Id, config); err != nil {
			return errors.NewInternalError("policy_failed", err.Error())
		}
	}
	price.PaymentReceiveRuleId = created.Id
	if _, err := p.paymentReceiveRulePriceRepository.Create(ctx, price); err != nil {
		return errors.NewInternalError("policy_failed", err.Error())
	}
	return nil
}

// defaultRule is a free price on the agent token for wallet and owner; nil without addresses.
func (p *PaymentReceivePolicyProvider) defaultRule(
	policy *models.PaymentReceivePolicy,
	agent *identity_models.Agent,
	evmAddress string,
) (*models.PaymentReceiveRule, json.RawMessage, *models.PaymentReceiveRulePrice, error) {
	patterns := []string{}
	if len(evmAddress) != 0 {
		patterns = append(patterns, strings.ToLower(evmAddress))
	}
	if agent.Owner != nil && len(*agent.Owner) != 0 {
		patterns = append(patterns, strings.ToLower(*agent.Owner))
	}
	if len(patterns) == 0 {
		return nil, nil, nil, nil
	}
	eip712Name := DefaultEip712Name
	eip712Version := DefaultEip712Version
	asset, _ := p.token6022Settings.AddressFor(lo.FromPtr(agent.ChainId))
	// x402 networks are CAIP-2; the agent stores a plain chain id.
	evmChainId, _ := strconv.ParseUint(lo.FromPtr(agent.ChainId), 10, 64)
	config := &models.PaymentReceiveRuleX402Config{
		Network:             helpers.FormatCaip2(evmChainId),
		Asset:               strings.ToLower(asset.Hex()),
		PayTo:               evmAddress,
		AssetTransferMethod: enums.AssetTransferMethodEip3009,
		Eip712Name:          &eip712Name,
		Eip712Version:       &eip712Version,
	}
	rawConfig, err := config.Marshal()
	if err != nil {
		return nil, nil, nil, err
	}
	return &models.PaymentReceiveRule{
		PolicyId: policy.Id,
		Name:     "Owner free access",
		Method:   enums.PaymentMethodX402,
	}, rawConfig, &models.PaymentReceiveRulePrice{
		Price:           "0",
		AddressPatterns: patterns,
	}, nil
}

func (p *PaymentReceivePolicyProvider) EvmAddress(ctx context.Context, agentId uuid.UUID) (string, error) {
	address, err := p.agentAddressRepository.GetByAgentIdAndType(ctx, agentId, identity_enums.AddressTypeEvm)
	if err != nil {
		return "", err
	}
	if address == nil {
		return "", nil
	}
	return address.Address, nil
}
