package interfaces

import (
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/google/uuid"
)

// PaymentReceiveMethodStrategyInterface owns one payment method: validates its Config blob and persists its config row.
type PaymentReceiveMethodStrategyInterface interface {
	Method() enums.PaymentMethod
	// Validate checks and sanitizes the request's Config blob in place, given the rule's prices.
	Validate(ctx context.Context, request *requests.SavePaymentReceiveRuleRequest) error
	// RuleIdForAsset returns the policy's rule already pricing the validated blob's token, or uuid.Nil.
	RuleIdForAsset(ctx context.Context, policyId uuid.UUID, config json.RawMessage) (uuid.UUID, error)
	// SaveConfig persists the rule's method config from the validated blob,
	// replacing any existing one.
	SaveConfig(ctx context.Context, policyId uuid.UUID, ruleId uuid.UUID, config json.RawMessage) error
	// LoadConfig returns the rule's method config as the wire JSON blob.
	LoadConfig(ctx context.Context, ruleId uuid.UUID) (json.RawMessage, error)
}
