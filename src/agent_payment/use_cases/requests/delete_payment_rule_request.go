package requests

import "github.com/google/uuid"

// DeletePaymentRuleRequest identifies the payment rule to delete, decoded from the route params.
type DeletePaymentRuleRequest struct {
	Id uuid.UUID `params:"id"`
}
