package services

import (
	"context"
	"encoding/json"
	"math/big"
	"regexp"
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
	// DefaultEip712Name/Version are the placeholder domain for the seeded free
	// rule, which never settles — operator rules derive the real domain
	// on-chain.
	DefaultEip712Name    = "USDC"
	DefaultEip712Version = "2"
)

// X402PaymentMethodStrategy is the x402 payment method: it validates a rule's
// config (must target a reachable chain, hex asset/payTo, address patterns, a
// real ERC-20 asset whose EIP-712 domain is derived on-chain, a facilitator for
// paid rules) and persists/loads it in the typed payment_receive_rule_x402_configs
// table. The settlement timeout is network-level (facilitator config), not here.
type X402PaymentMethodStrategy struct {
	paymentReceiveSettings                 *runtime_common_settings.PaymentReceiveSettings
	evmChainsSettings                      *runtime_common_settings.EvmChainsSettings
	erc20Validator                         payment_interfaces.Erc20ValidatorInterface
	paymentReceiveRuleX402ConfigRepository repositories.PaymentReceiveRuleX402ConfigRepositoryInterface
}

func NewX402PaymentMethodStrategy(
	paymentReceiveSettings *runtime_common_settings.PaymentReceiveSettings,
	evmChainsSettings *runtime_common_settings.EvmChainsSettings,
	erc20Validator payment_interfaces.Erc20ValidatorInterface,
	paymentReceiveRuleX402ConfigRepository repositories.PaymentReceiveRuleX402ConfigRepositoryInterface,
) *X402PaymentMethodStrategy {
	return &X402PaymentMethodStrategy{
		paymentReceiveSettings:                 paymentReceiveSettings,
		evmChainsSettings:                      evmChainsSettings,
		erc20Validator:                         erc20Validator,
		paymentReceiveRuleX402ConfigRepository: paymentReceiveRuleX402ConfigRepository,
	}
}

func (s *X402PaymentMethodStrategy) Method() enums.PaymentMethod {
	return enums.PaymentMethodX402
}

func (s *X402PaymentMethodStrategy) Validate(
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
	chainId := strconv.FormatUint(parsedChainId, 10)
	// Rules can only target a reachable chain (needed to validate the asset).
	if !s.evmChainsSettings.Has(chainId) {
		return errors.NewValidationError("config.network", "node has no RPC for this network")
	}
	if !common.IsHexAddress(config.Asset) {
		return errors.NewValidationError("config.asset", "must be a hex address")
	}
	if !common.IsHexAddress(config.PayTo) {
		return errors.NewValidationError("config.payTo", "must be a hex address")
	}
	if len(config.AddressPatterns) == 0 {
		return errors.NewValidationError(
			"config.addressPatterns",
			"at least one pattern is required (\".*\" opens the rule to everyone)",
		)
	}
	for i, pattern := range config.AddressPatterns {
		trimmed := strings.TrimSpace(pattern)
		if len(trimmed) == 0 {
			return errors.NewValidationError("config.addressPatterns", "patterns cannot be empty")
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return errors.NewValidationError(
				"config.addressPatterns",
				"must be valid regexes matched against the lowercase 0x address",
			)
		}
		config.AddressPatterns[i] = trimmed
	}

	// The transfer method + EIP-712 domain belong to the token, so detect them
	// on-chain rather than trust the operator — a wrong domain would silently
	// fail settlement.
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

	if err := s.requireFacilitatorForPaidRule(request.Price, config.Network); err != nil {
		return err
	}

	sanitized, err := config.Marshal()
	if err != nil {
		return errors.NewValidationError("config", "could not be encoded")
	}
	request.Config = sanitized
	return nil
}

func (s *X402PaymentMethodStrategy) SaveConfig(ctx context.Context, ruleId uuid.UUID, raw json.RawMessage) error {
	config := &models.PaymentReceiveRuleX402Config{}
	if err := config.Parse(raw); err != nil {
		return err
	}
	config.PaymentReceiveRuleId = ruleId
	return s.paymentReceiveRuleX402ConfigRepository.Upsert(ctx, config)
}

func (s *X402PaymentMethodStrategy) LoadConfig(ctx context.Context, ruleId uuid.UUID) (json.RawMessage, error) {
	row, err := s.paymentReceiveRuleX402ConfigRepository.GetByRuleId(ctx, ruleId)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return row.Marshal()
}

// requireFacilitatorForPaidRule rejects a paid rule whose network has no
// configured facilitator. Free rules never settle, so they're exempt.
func (s *X402PaymentMethodStrategy) requireFacilitatorForPaidRule(price string, network string) error {
	parsed, ok := new(big.Int).SetString(price, 10)
	if !ok || parsed.Sign() == 0 {
		return nil
	}
	if _, configured := s.paymentReceiveSettings.BaseUrlFor(network); !configured {
		return errors.NewValidationError("config.network", "no facilitator configured for this network")
	}
	return nil
}
