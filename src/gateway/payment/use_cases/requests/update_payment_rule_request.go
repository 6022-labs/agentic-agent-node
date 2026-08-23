package requests

import "github.com/google/uuid"

// UpdatePaymentRuleRequest pairs the route id with the replacement rule body.
type UpdatePaymentRuleRequest struct {
	Id uuid.UUID `uri:"id" json:"-"`
	SavePaymentRuleRequest
}
