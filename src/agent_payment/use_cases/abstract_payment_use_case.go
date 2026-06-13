package use_cases

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

// AbstractPaymentUseCase carries the repositories every payment use case
// needs and the shared policy-materialization behavior.
type AbstractPaymentUseCase struct {
	agentRepository             identity_repositories.AgentRepositoryInterface
	paymentPolicyRepository     repositories.PaymentPolicyRepositoryInterface
	paymentRuleRepository       repositories.PaymentRuleRepositoryInterface
	blockchainSettings          *runtime_common_settings.BlockchainSettings
	token6022Settings           *runtime_common_settings.Token6022Settings
	paymentFacilitatorsSettings *runtime_common_settings.PaymentFacilitatorsSettings
}

func NewAbstractPaymentUseCase(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
	blockchainSettings *runtime_common_settings.BlockchainSettings,
	token6022Settings *runtime_common_settings.Token6022Settings,
	paymentFacilitatorsSettings *runtime_common_settings.PaymentFacilitatorsSettings,
) *AbstractPaymentUseCase {
	return &AbstractPaymentUseCase{
		agentRepository:             agentRepository,
		paymentPolicyRepository:     paymentPolicyRepository,
		paymentRuleRepository:       paymentRuleRepository,
		blockchainSettings:          blockchainSettings,
		token6022Settings:           token6022Settings,
		paymentFacilitatorsSettings: paymentFacilitatorsSettings,
	}
}

// requireFacilitatorForPaidRule rejects a paid rule whose network has no
// configured facilitator. Free rules never settle, so they're exempt.
func (u *AbstractPaymentUseCase) requireFacilitatorForPaidRule(request *requests.SavePaymentRuleRequest) error {
	price, ok := new(big.Int).SetString(request.Price, 10)
	if !ok || price.Sign() == 0 {
		return nil
	}
	if _, configured := u.paymentFacilitatorsSettings.BaseUrlFor(request.Network); !configured {
		return errors.NewValidationError("network", "no facilitator configured for this network")
	}
	return nil
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
	if rule := u.defaultRule(policy, agent); rule != nil {
		if _, err := u.paymentRuleRepository.Create(ctx, rule); err != nil {
			return nil, nil, errors.NewInternalError("policy_failed", err.Error())
		}
	}
	return policy, agent, nil
}

// defaultRule seeds a zero-price route to the agent's own wallet, scoped to
// agent wallet + owner, on the node's chain. The $6022 token is only the
// EIP-712 signing context — zero-price rules never settle on-chain.
func (u *AbstractPaymentUseCase) defaultRule(policy *models.PaymentPolicy, agent *identity_models.Agent) *models.PaymentRule {
	patterns := []string{}
	if len(agent.WalletAddress) != 0 {
		patterns = append(patterns, strings.ToLower(agent.WalletAddress))
	}
	if agent.Owner != nil && len(*agent.Owner) != 0 {
		patterns = append(patterns, strings.ToLower(*agent.Owner))
	}
	if len(patterns) == 0 {
		return nil
	}
	return &models.PaymentRule{
		PolicyId:          policy.Id,
		Name:              "Owner free access",
		Network:           fmt.Sprintf("eip155:%d", u.blockchainSettings.Id),
		Asset:             u.token6022Settings.SmartContractAddress.Hex(),
		PayTo:             agent.WalletAddress,
		Price:             "0",
		MaxTimeoutSeconds: 60,
		Eip712Name:        "USDC",
		Eip712Version:     "2",
		AddressPatterns:   patterns,
	}
}
