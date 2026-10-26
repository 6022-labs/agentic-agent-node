package requests

import "github.com/google/uuid"

// DeletePaymentReceiveRuleRequest identifies the payment rule to delete, decoded from the route params.
type DeletePaymentReceiveRuleRequest struct {
	Id uuid.UUID `uri:"id"`
}
