package requests

import "github.com/google/uuid"

// UpdatePaymentReceiveRuleRequest pairs the route id with the replacement rule body.
type UpdatePaymentReceiveRuleRequest struct {
	Id uuid.UUID `uri:"id" json:"-"`
	SavePaymentReceiveRuleRequest
}
