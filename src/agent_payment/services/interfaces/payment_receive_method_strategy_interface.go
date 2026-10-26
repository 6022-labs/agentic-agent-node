package interfaces

import (
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/google/uuid"
)

// PaymentReceiveMethodStrategyInterface owns one payment method: it validates the
// request's Config blob and persists/loads that method's typed config row. The
// generic payment domain dispatches to it by the rule's Method discriminator,
// so adding a method (e.g. Lightning) is a new strategy + its own config table,
// nothing in the generic layer.
type PaymentReceiveMethodStrategyInterface interface {
	Method() enums.PaymentMethod
	// Validate checks and sanitizes the request's Config blob in place.
	Validate(ctx context.Context, request *requests.SavePaymentReceiveRuleRequest) error
	// SaveConfig persists the rule's method config from the validated blob,
	// replacing any existing one.
	SaveConfig(ctx context.Context, ruleId uuid.UUID, config json.RawMessage) error
	// LoadConfig returns the rule's method config as the wire JSON blob.
	LoadConfig(ctx context.Context, ruleId uuid.UUID) (json.RawMessage, error)
}
