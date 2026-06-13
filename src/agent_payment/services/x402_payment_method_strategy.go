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
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

const (
	caip2Eip155Prefix = "eip155:"
	// DefaultEip712Name/Version are the USDC EIP-712 domain defaults applied
	// when a rule's config omits them.
	DefaultEip712Name    = "USDC"
	DefaultEip712Version = "2"
)

// X402PaymentMethodStrategy is the x402 payment method: it validates a rule's
// config (CAIP-2 network, hex asset/payTo, address patterns, EIP-712 defaults,
// a real ERC-20 asset, a facilitator for paid rules) and persists/loads it in
// the typed payment_rule_x402_configs table. The settlement timeout is
// network-level, so it lives in the facilitator config, not here.
type X402PaymentMethodStrategy struct {
	paymentFacilitatorsSettings     *runtime_common_settings.PaymentFacilitatorsSettings
	erc20Validator                  payment_interfaces.Erc20ValidatorInterface
	paymentRuleX402ConfigRepository repositories.PaymentRuleX402ConfigRepositoryInterface
}

func NewX402PaymentMethodStrategy(
	paymentFacilitatorsSettings *runtime_common_settings.PaymentFacilitatorsSettings,
	erc20Validator payment_interfaces.Erc20ValidatorInterface,
	paymentRuleX402ConfigRepository repositories.PaymentRuleX402ConfigRepositoryInterface,
) *X402PaymentMethodStrategy {
	return &X402PaymentMethodStrategy{
		paymentFacilitatorsSettings:     paymentFacilitatorsSettings,
		erc20Validator:                  erc20Validator,
		paymentRuleX402ConfigRepository: paymentRuleX402ConfigRepository,
	}
}

func (s *X402PaymentMethodStrategy) Method() enums.PaymentMethod {
	return enums.PaymentMethodX402
}

func (s *X402PaymentMethodStrategy) Validate(ctx context.Context, request *requests.SavePaymentRuleRequest) error {
	config := &models.PaymentRuleX402Config{}
	if err := config.Parse(request.Config); err != nil {
		return err
	}

	config.Network = strings.TrimSpace(config.Network)
	config.Asset = strings.TrimSpace(config.Asset)
	config.PayTo = strings.TrimSpace(config.PayTo)
	config.Eip712Name = strings.TrimSpace(config.Eip712Name)
	config.Eip712Version = strings.TrimSpace(config.Eip712Version)

	chainId, err := s.parseEip155ChainId(config.Network)
	if err != nil {
		return err
	}
	if !common.IsHexAddress(config.Asset) {
		return errors.NewValidationError("config.asset", "must be a hex address")
	}
	if !common.IsHexAddress(config.PayTo) {
		return errors.NewValidationError("config.payTo", "must be a hex address")
	}
	if len(config.AddressPatterns) == 0 {
		return errors.NewValidationError("config.addressPatterns", "at least one pattern is required (\".*\" opens the rule to everyone)")
	}
	for i, pattern := range config.AddressPatterns {
		trimmed := strings.TrimSpace(pattern)
		if len(trimmed) == 0 {
			return errors.NewValidationError("config.addressPatterns", "patterns cannot be empty")
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return errors.NewValidationError("config.addressPatterns", "must be valid regexes matched against the lowercase 0x address")
		}
		config.AddressPatterns[i] = trimmed
	}

	if len(config.Eip712Name) == 0 {
		config.Eip712Name = DefaultEip712Name
	}
	if len(config.Eip712Version) == 0 {
		config.Eip712Version = DefaultEip712Version
	}

	if err := s.erc20Validator.ValidateErc20(ctx, chainId, config.Asset); err != nil {
		return errors.NewValidationError("config.asset", err.Error())
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
	config := &models.PaymentRuleX402Config{}
	if err := config.Parse(raw); err != nil {
		return err
	}
	config.PaymentRuleId = ruleId
	return s.paymentRuleX402ConfigRepository.Upsert(ctx, config)
}

func (s *X402PaymentMethodStrategy) LoadConfig(ctx context.Context, ruleId uuid.UUID) (json.RawMessage, error) {
	row, err := s.paymentRuleX402ConfigRepository.GetByRuleId(ctx, ruleId)
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
	if _, configured := s.paymentFacilitatorsSettings.BaseUrlFor(network); !configured {
		return errors.NewValidationError("config.network", "no facilitator configured for this network")
	}
	return nil
}

// parseEip155ChainId extracts the chain id from a CAIP-2 eip155 network
// identifier ("eip155:80002" → 80002).
func (s *X402PaymentMethodStrategy) parseEip155ChainId(network string) (int64, error) {
	rest, found := strings.CutPrefix(network, caip2Eip155Prefix)
	if !found {
		return 0, errors.NewValidationError("config.network", "must be a CAIP-2 eip155 identifier, e.g. eip155:80002")
	}
	chainId, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || chainId <= 0 {
		return 0, errors.NewValidationError("config.network", "must be a CAIP-2 eip155 identifier, e.g. eip155:80002")
	}
	return chainId, nil
}
