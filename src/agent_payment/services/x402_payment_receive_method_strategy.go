package services

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

const (
	// DefaultEip712Name/Version are the placeholder domain for the seeded free rule, which never settles.
	DefaultEip712Name    = "USDC"
	DefaultEip712Version = "2"
)

// X402PaymentReceiveMethodStrategy validates a rule's config and persists it in payment_receive_rule_x402_configs.
type X402PaymentReceiveMethodStrategy struct {
	paymentFacilitatorsSettings            *runtime_common_settings.PaymentReceiveSettings
	evmChainsSettings                      *runtime_common_settings.EvmChainsSettings
	erc20Validator                         payment_interfaces.Erc20ValidatorInterface
	paymentReceiveRuleX402ConfigRepository repositories.PaymentReceiveRuleX402ConfigRepositoryInterface
}

func NewX402PaymentReceiveMethodStrategy(
	paymentFacilitatorsSettings *runtime_common_settings.PaymentReceiveSettings,
	evmChainsSettings *runtime_common_settings.EvmChainsSettings,
	erc20Validator payment_interfaces.Erc20ValidatorInterface,
	paymentReceiveRuleX402ConfigRepository repositories.PaymentReceiveRuleX402ConfigRepositoryInterface,
) *X402PaymentReceiveMethodStrategy {
	return &X402PaymentReceiveMethodStrategy{
		paymentFacilitatorsSettings:            paymentFacilitatorsSettings,
		evmChainsSettings:                      evmChainsSettings,
		erc20Validator:                         erc20Validator,
		paymentReceiveRuleX402ConfigRepository: paymentReceiveRuleX402ConfigRepository,
	}
}

func (s *X402PaymentReceiveMethodStrategy) Method() enums.PaymentMethod {
	return enums.PaymentMethodX402
}

func (s *X402PaymentReceiveMethodStrategy) Validate(
	ctx context.Context,
	request *requests.SavePaymentReceiveRuleRequest,
) error {
	config := &models.PaymentReceiveRuleX402Config{}
	if err := config.Parse(request.Config); err != nil {
		return err
	}

	config.Network = strings.TrimSpace(config.Network)
	config.Asset = strings.TrimSpace(config.Asset)
	config.PayTo = strings.TrimSpace(config.PayTo)

	// The x402 network is CAIP-2; map it to the plain chain id.
	parsedChainId, err := helpers.ParseCaip2(config.Network)
	if err != nil {
		return errors.NewValidationError("config.network", err.Error())
	}
	// Canonical: part of the one-rule-per-token key.
	config.Network = helpers.FormatCaip2(parsedChainId)
	chainId := strconv.FormatUint(parsedChainId, 10)
	// Rules can only target a reachable chain (needed to validate the asset).
	if !s.evmChainsSettings.Has(chainId) {
		return errors.NewValidationError("config.network", "node has no RPC for this network")
	}
	if !common.IsHexAddress(config.Asset) {
		return errors.NewValidationError("config.asset", "must be a hex address")
	}
	// Lowercased: part of the one-rule-per-token key.
	config.Asset = strings.ToLower(config.Asset)
	if !common.IsHexAddress(config.PayTo) {
		return errors.NewValidationError("config.payTo", "must be a hex address")
	}

	// Detect the transfer method + domain on-chain; a wrong domain silently fails settlement.
	method, eip712Name, eip712Version, err := s.erc20Validator.ValidateErc20(ctx, chainId, config.Asset)
	if err != nil {
		return errors.NewValidationError("config.asset", err.Error())
	}
	config.AssetTransferMethod = method
	// Only eip3009 carries the token's EIP-712 domain; permit2 leaves it nil.
	if method == enums.AssetTransferMethodEip3009 {
		config.Eip712Name = &eip712Name
		config.Eip712Version = &eip712Version
	} else {
		config.Eip712Name = nil
		config.Eip712Version = nil
	}

	if request.HasPaidPrice() {
		if _, configured := s.paymentFacilitatorsSettings.BaseUrlFor(config.Network); !configured {
			return errors.NewValidationError("config.network", "no facilitator configured for this network")
		}
	}

	sanitized, err := config.Marshal()
	if err != nil {
		return errors.NewValidationError("config", "could not be encoded")
	}
	request.Config = sanitized
	return nil
}

func (s *X402PaymentReceiveMethodStrategy) RuleIdForAsset(
	ctx context.Context,
	policyId uuid.UUID,
	raw json.RawMessage,
) (uuid.UUID, error) {
	config := &models.PaymentReceiveRuleX402Config{}
	if err := config.Parse(raw); err != nil {
		return uuid.Nil, err
	}
	existing, err := s.paymentReceiveRuleX402ConfigRepository.GetByPolicyIdAndAsset(
		ctx,
		policyId,
		config.Network,
		config.Asset,
	)
	if err != nil {
		return uuid.Nil, err
	}
	if existing == nil {
		return uuid.Nil, nil
	}
	return existing.PaymentReceiveRuleId, nil
}

func (s *X402PaymentReceiveMethodStrategy) SaveConfig(
	ctx context.Context,
	policyId uuid.UUID,
	ruleId uuid.UUID,
	raw json.RawMessage,
) error {
	config := &models.PaymentReceiveRuleX402Config{}
	if err := config.Parse(raw); err != nil {
		return err
	}
	config.PaymentReceiveRuleId = ruleId
	config.PolicyId = policyId
	return s.paymentReceiveRuleX402ConfigRepository.Upsert(ctx, config)
}
func (s *X402PaymentReceiveMethodStrategy) LoadConfig(ctx context.Context, ruleId uuid.UUID) (json.RawMessage, error) {
	row, err := s.paymentReceiveRuleX402ConfigRepository.GetByRuleId(ctx, ruleId)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return row.Marshal()
}
