package use_cases

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

// AbstractPaymentUseCase carries the repositories every payment use case
// needs and the shared policy-materialization behavior.
type AbstractPaymentUseCase struct {
	agentRepository         identity_repositories.AgentRepositoryInterface
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface
	paymentRuleRepository   repositories.PaymentRuleRepositoryInterface
	paymentMethodStrategies payment_interfaces.PaymentMethodStrategiesInterface
	token6022Settings       *runtime_common_settings.Token6022Settings
}

func NewAbstractPaymentUseCase(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
	paymentMethodStrategies payment_interfaces.PaymentMethodStrategiesInterface,
	token6022Settings *runtime_common_settings.Token6022Settings,
) *AbstractPaymentUseCase {
	return &AbstractPaymentUseCase{
		agentRepository:         agentRepository,
		paymentPolicyRepository: paymentPolicyRepository,
		paymentRuleRepository:   paymentRuleRepository,
		paymentMethodStrategies: paymentMethodStrategies,
		token6022Settings:       token6022Settings,
	}
}

// ensurePaymentPolicy returns the policy head, materializing it with the
// seeded default rule on first access.
func (u *AbstractPaymentUseCase) ensurePaymentPolicy(ctx context.Context) (*models.PaymentPolicy, *identity_models.Agent, error) {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, nil, errors.NewInternalError("policy_failed", err.Error())
	}
	if agent == nil {
		return nil, nil, errors.NewInternalError("policy_failed", "agent identity is not initialized yet")
	}

	policy, err := u.paymentPolicyRepository.Get(ctx)
	if err != nil {
		return nil, nil, errors.NewInternalError("policy_failed", err.Error())
	}
	if policy != nil {
		return policy, agent, nil
	}

	policy, err = u.paymentPolicyRepository.Create(ctx, &models.PaymentPolicy{AgentId: agent.Id})
	if err != nil {
		return nil, nil, errors.NewInternalError("policy_failed", err.Error())
	}
	if rule, config := u.defaultRule(policy, agent); rule != nil {
		created, err := u.paymentRuleRepository.Create(ctx, rule)
		if err != nil {
			return nil, nil, errors.NewInternalError("policy_failed", err.Error())
		}
		if strategy, ok := u.paymentMethodStrategies.For(created.Method); ok {
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
func (u *AbstractPaymentUseCase) defaultRule(policy *models.PaymentPolicy, agent *identity_models.Agent) (*models.PaymentRule, json.RawMessage) {
	patterns := []string{}
	if len(agent.WalletAddress) != 0 {
		patterns = append(patterns, strings.ToLower(agent.WalletAddress))
	}
	if agent.Owner != nil && len(*agent.Owner) != 0 {
		patterns = append(patterns, strings.ToLower(*agent.Owner))
	}
	if len(patterns) == 0 {
		return nil, nil
	}
	eip712Name := services.DefaultEip712Name
	eip712Version := services.DefaultEip712Version
	asset, _ := u.token6022Settings.AddressFor(agent.ChainId)
	// x402 networks are CAIP-2; the agent stores a plain chain id.
	evmChainId, _ := strconv.ParseUint(agent.ChainId, 10, 64)
	config := &models.PaymentRuleX402Config{
		Network:             helpers.FormatCaip2(evmChainId),
		Asset:               asset.Hex(),
		PayTo:               agent.WalletAddress,
		AssetTransferMethod: enums.AssetTransferMethodEip3009,
		Eip712Name:          &eip712Name,
		Eip712Version:       &eip712Version,
		AddressPatterns:     patterns,
	}
	rawConfig, err := config.Marshal()
	if err != nil {
		return nil, nil
	}
	return &models.PaymentRule{
		PolicyId: policy.Id,
		Name:     "Owner free access",
		Price:    "0",
		Method:   enums.PaymentMethodX402,
	}, rawConfig
}
