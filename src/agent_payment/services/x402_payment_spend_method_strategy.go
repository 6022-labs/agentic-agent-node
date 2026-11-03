package services

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// X402PaymentSpendMethodStrategy validates a spend rule's token; unlike receive it touches no chain.
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
	chainId, err := helpers.ParseCaip2(config.Network)
	if err != nil {
		return errors.NewValidationError("config.network", err.Error())
	}
	// Canonical because it is part of the one-rule-per-token key.
	config.Network = helpers.FormatCaip2(chainId)
	// The 0x prefix is required, as on the quote, or the rule could never match one.
	if !helpers.IsEthAddress(config.Asset) {
		return errors.NewValidationError("config.asset", "must be the 0x-prefixed token contract address")
	}
	// Lowercased because it is the lookup key: every pasted case must store alike.
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

	raw, err := config.Marshal()
	if err != nil {
		return errors.NewValidationError("config", "could not be encoded")
	}
	request.Config = raw
	return nil
}

func (s *X402PaymentSpendMethodStrategy) RuleIdForAsset(
	ctx context.Context,
	policyId uuid.UUID,
	raw json.RawMessage,
) (uuid.UUID, error) {
	config := &models.PaymentSpendRuleX402Config{}
	if err := config.Parse(raw); err != nil {
		return uuid.Nil, err
	}
	existing, err := s.paymentSpendRuleX402ConfigRepository.GetByPolicyIdAndAsset(
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
	return existing.PaymentSpendRuleId, nil
}

func (s *X402PaymentSpendMethodStrategy) SaveConfig(
	ctx context.Context,
	policyId uuid.UUID,
	ruleId uuid.UUID,
	raw json.RawMessage,
) error {
	config := &models.PaymentSpendRuleX402Config{}
	if err := config.Parse(raw); err != nil {
		return err
	}
	config.PaymentSpendRuleId = ruleId
	config.PolicyId = policyId
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
