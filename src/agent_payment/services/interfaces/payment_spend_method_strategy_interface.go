package interfaces

import (
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/google/uuid"
)

// PaymentSpendMethodStrategyInterface owns one payment method on the spend side:
// it validates the rule's Config blob and persists/loads that method's matcher.
type PaymentSpendMethodStrategyInterface interface {
	Method() enums.PaymentMethod
	// Validate checks and sanitizes the request's Config blob in place.
	Validate(ctx context.Context, request *requests.SavePaymentSpendRuleRequest) error
	// SaveConfig persists the rule's matcher from the validated blob, replacing any existing one.
	SaveConfig(ctx context.Context, ruleId uuid.UUID, config json.RawMessage) error
	// LoadConfig returns the rule's matcher as the wire JSON blob.
	LoadConfig(ctx context.Context, ruleId uuid.UUID) (json.RawMessage, error)
}
