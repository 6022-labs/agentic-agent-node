package services

import (
	"context"
	"encoding/json"
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

	facets := []struct {
		field    string
		patterns *[]string
	}{
		{"networkPatterns", &config.NetworkPatterns},
		{"assetPatterns", &config.AssetPatterns},
		{"payToPatterns", &config.PayToPatterns},
		{"assetTransferMethodPatterns", &config.AssetTransferMethodPatterns},
	}
	for _, facet := range facets {
		sanitized, err := s.sanitizePatterns(facet.field, *facet.patterns)
		if err != nil {
			return err
		}
		*facet.patterns = sanitized
	}

	raw, err := config.Marshal()
	if err != nil {
		return errors.NewValidationError("config", "could not be encoded")
	}
	request.Config = raw
	return nil
}

// sanitizePatterns trims and compiles each pattern; an empty array would match nothing, denying the facet outright.
func (s *X402PaymentSpendMethodStrategy) sanitizePatterns(field string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, errors.NewValidationError(
			"config."+field,
			"at least one pattern is required (\".*\" allows every value)",
		)
	}

	sanitized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		trimmed := strings.TrimSpace(pattern)
		if len(trimmed) == 0 {
			return nil, errors.NewValidationError("config."+field, "patterns cannot be empty")
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return nil, errors.NewValidationError("config."+field, "must be valid regexes")
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
