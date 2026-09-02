package requests

import "github.com/google/uuid"

// DeletePaymentSpendRuleRequest identifies the spend rule to delete, decoded from the route params.
type DeletePaymentSpendRuleRequest struct {
	Id uuid.UUID `uri:"id"`
}
