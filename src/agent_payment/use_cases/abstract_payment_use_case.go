package use_cases

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
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// AbstractPaymentUseCase carries the repositories every payment use case
// needs and the shared policy-materialization behavior.
type AbstractPaymentUseCase struct {
	agentRepository                identity_repositories.AgentRepositoryInterface
	agentAddressRepository         identity_repositories.AgentAddressRepositoryInterface
	paymentReceivePolicyRepository repositories.PaymentReceivePolicyRepositoryInterface
	paymentReceiveRuleRepository   repositories.PaymentReceiveRuleRepositoryInterface
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface
	token6022Settings              *runtime_common_settings.Token6022Settings
}

func NewAbstractPaymentUseCase(
	agentRepository identity_repositories.AgentRepositoryInterface,
	agentAddressRepository identity_repositories.AgentAddressRepositoryInterface,
	paymentReceivePolicyRepository repositories.PaymentReceivePolicyRepositoryInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface,
	token6022Settings *runtime_common_settings.Token6022Settings,
) *AbstractPaymentUseCase {
	return &AbstractPaymentUseCase{
		agentRepository:                agentRepository,
		agentAddressRepository:         agentAddressRepository,
		paymentReceivePolicyRepository: paymentReceivePolicyRepository,
		paymentReceiveRuleRepository:   paymentReceiveRuleRepository,
		paymentReceiveMethodStrategies: paymentReceiveMethodStrategies,
		token6022Settings:              token6022Settings,
	}
}

// ensurePaymentReceivePolicy returns the policy head, materializing it with the
// seeded default rule on first access.
func (u *AbstractPaymentUseCase) ensurePaymentReceivePolicy(
	ctx context.Context,
) (*models.PaymentReceivePolicy, *identity_models.Agent, error) {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, nil, errors.NewInternalError("policy_failed", err.Error())
	}
	if agent == nil {
		return nil, nil, errors.NewInternalError("policy_failed", "agent identity is not initialized yet")
	}
	// Do not create default payment policy if agent is not minted
	if agent.CollectionTokenId == nil {
		return nil, nil, errors.NewConflictError("agent_not_minted", "agent identity is not minted yet")
	}

	policy, err := u.paymentReceivePolicyRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, nil, errors.NewInternalError("policy_failed", err.Error())
	}
	if policy != nil {
		return policy, agent, nil
	}

	policy, err = u.paymentReceivePolicyRepository.Create(ctx, &models.PaymentReceivePolicy{AgentId: agent.Id})
	if err != nil {
		// A concurrent first request may have won the insert race; re-read before failing.
		existing, getErr := u.paymentReceivePolicyRepository.GetByAgentId(ctx, agent.Id)
		if getErr == nil && existing != nil {
			return existing, agent, nil
		}
		return nil, nil, errors.NewInternalError("policy_failed", err.Error())
	}
	evmAddress, err := u.evmAddress(ctx, agent.Id)
	if err != nil {
		return nil, nil, errors.NewInternalError("policy_failed", err.Error())
	}
	if rule, config := u.defaultRule(policy, agent, evmAddress); rule != nil {
		created, err := u.paymentReceiveRuleRepository.Create(ctx, rule)
		if err != nil {
			return nil, nil, errors.NewInternalError("policy_failed", err.Error())
		}
		if strategy, ok := u.paymentReceiveMethodStrategies.For(created.Method); ok {
			if err := strategy.SaveConfig(ctx, created.Id, config); err != nil {
				return nil, nil, errors.NewInternalError("policy_failed", err.Error())
			}
		}
	}
	return policy, agent, nil
}

// defaultRule seeds a zero-price x402 rule routing to the agent's own wallet,
// scoped to agent wallet + owner on the agent's identity chain, and returns it
// with the config JSON the x402 strategy persists. The $6022 token is only the
// EIP-712 signing context — zero-price rules never settle on-chain. Returns
// (nil, nil) when the agent has no addresses to scope to.
func (u *AbstractPaymentUseCase) defaultRule(
	policy *models.PaymentReceivePolicy,
	agent *identity_models.Agent,
	evmAddress string,
) (*models.PaymentReceiveRule, json.RawMessage) {
	patterns := []string{}
	if len(evmAddress) != 0 {
		patterns = append(patterns, strings.ToLower(evmAddress))
	}
	if agent.Owner != nil && len(*agent.Owner) != 0 {
		patterns = append(patterns, strings.ToLower(*agent.Owner))
	}
	if len(patterns) == 0 {
		return nil, nil
	}
	eip712Name := services.DefaultEip712Name
	eip712Version := services.DefaultEip712Version
	asset, _ := u.token6022Settings.AddressFor(lo.FromPtr(agent.ChainId))
	// x402 networks are CAIP-2; the agent stores a plain chain id.
	evmChainId, _ := strconv.ParseUint(lo.FromPtr(agent.ChainId), 10, 64)
	config := &models.PaymentReceiveRuleX402Config{
		Network:             helpers.FormatCaip2(evmChainId),
		Asset:               asset.Hex(),
		PayTo:               evmAddress,
		AssetTransferMethod: enums.AssetTransferMethodEip3009,
		Eip712Name:          &eip712Name,
		Eip712Version:       &eip712Version,
		AddressPatterns:     patterns,
	}
	rawConfig, err := config.Marshal()
	if err != nil {
		return nil, nil
	}
	return &models.PaymentReceiveRule{
		PolicyId: policy.Id,
		Name:     "Owner free access",
		Price:    "0",
		Method:   enums.PaymentMethodX402,
	}, rawConfig
}

// evmAddress returns the agent's stored evm address, or "" when none exists.
func (u *AbstractPaymentUseCase) evmAddress(ctx context.Context, agentId uuid.UUID) (string, error) {
	address, err := u.agentAddressRepository.GetByAgentIdAndType(ctx, agentId, identity_enums.AddressTypeEvm)
	if err != nil {
		return "", err
	}
	if address == nil {
		return "", nil
	}
	return address.Address, nil
}
