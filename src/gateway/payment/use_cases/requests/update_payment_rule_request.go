package requests

import "github.com/google/uuid"

// UpdatePaymentRuleRequest pairs the route id with the replacement rule body.
type UpdatePaymentRuleRequest struct {
	Id uuid.UUID `params:"id" json:"-"`
	SavePaymentRuleRequest
}
