package requests

import "github.com/google/uuid"

// UpdatePaymentSpendRuleRequest pairs the route id with the replacement rule body.
type UpdatePaymentSpendRuleRequest struct {
	Id uuid.UUID `uri:"id" json:"-"`
	SavePaymentSpendRuleRequest
}
