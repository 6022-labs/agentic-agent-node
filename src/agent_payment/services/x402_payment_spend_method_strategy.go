package services

import (
	"context"
	"encoding/json"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/ethereum/go-ethereum/common"
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// X402PaymentSpendMethodStrategy validates a spend rule's matcher; unlike receive it touches no chain.
type X402PaymentSpendMethodStrategy struct {
	paymentSpendRuleX402ConfigRepository repositories.PaymentSpendRuleX402ConfigRepositoryInterface
}

func NewX402PaymentSpendMethodStrategy(
	paymentSpendRuleX402ConfigRepository repositories.PaymentSpendRuleX402ConfigRepositoryInterface,
) *X402PaymentSpendMethodStrategy {
	return &X402PaymentSpendMethodStrategy{
		paymentSpendRuleX402ConfigRepository: paymentSpendRuleX402ConfigRepository,
	}
}

func (s *X402PaymentSpendMethodStrategy) Method() enums.PaymentMethod {
	return enums.PaymentMethodX402
}

func (s *X402PaymentSpendMethodStrategy) Validate(
	_ context.Context,
	request *requests.SavePaymentSpendRuleRequest,
) error {
	config := &models.PaymentSpendRuleX402Config{}
	if err := config.Parse(request.Config); err != nil {
		return err
	}

	config.Network = strings.TrimSpace(config.Network)
	config.Asset = strings.TrimSpace(config.Asset)

	// The x402 network is CAIP-2.
	if _, err := helpers.ParseCaip2(config.Network); err != nil {
		return errors.NewValidationError("config.network", err.Error())
	}
	if !common.IsHexAddress(config.Asset) {
		return errors.NewValidationError("config.asset", "must be the token contract address")
	}
	// Lowercased so the stored value reads the same whatever case was pasted.
	config.Asset = strings.ToLower(config.Asset)

	if config.AssetTransferMethod != nil {
		var method enums.AssetTransferMethod
		if err := method.Set(config.AssetTransferMethod.String()); err != nil {
			return errors.NewValidationError(
				"config.assetTransferMethod",
				"must be one of eip3009, permit2, or omitted",
			)
		}
		config.AssetTransferMethod = &method
	}

	sanitized, err := s.sanitizePayToPatterns(config.PayToPatterns)
	if err != nil {
		return err
	}
	config.PayToPatterns = sanitized

	raw, err := config.Marshal()
	if err != nil {
		return errors.NewValidationError("config", "could not be encoded")
	}
	request.Config = raw
	return nil
}

// sanitizePayToPatterns rejects an empty array, which would match nobody.
func (s *X402PaymentSpendMethodStrategy) sanitizePayToPatterns(patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, errors.NewValidationError(
			"config.payToPatterns",
			"at least one pattern is required (\".*\" allows every payee)",
		)
	}

	sanitized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		trimmed := strings.TrimSpace(pattern)
		if len(trimmed) == 0 {
			return nil, errors.NewValidationError("config.payToPatterns", "patterns cannot be empty")
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return nil, errors.NewValidationError(
				"config.payToPatterns",
				"must be valid regexes matched against the payee address",
			)
		}
		// A pasted checksummed address would compile and then never match the lowercased payee.
		if common.IsHexAddress(trimmed) {
			trimmed = strings.ToLower(trimmed)
		}
		sanitized = append(sanitized, trimmed)
	}
	return sanitized, nil
}

func (s *X402PaymentSpendMethodStrategy) SaveConfig(
	ctx context.Context,
	ruleId uuid.UUID,
	raw json.RawMessage,
) error {
	config := &models.PaymentSpendRuleX402Config{}
	if err := config.Parse(raw); err != nil {
		return err
	}
	config.PaymentSpendRuleId = ruleId
	return s.paymentSpendRuleX402ConfigRepository.Upsert(ctx, config)
}

func (s *X402PaymentSpendMethodStrategy) LoadConfig(
	ctx context.Context,
	ruleId uuid.UUID,
) (json.RawMessage, error) {
	row, err := s.paymentSpendRuleX402ConfigRepository.GetByRuleId(ctx, ruleId)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return row.Marshal()
}
