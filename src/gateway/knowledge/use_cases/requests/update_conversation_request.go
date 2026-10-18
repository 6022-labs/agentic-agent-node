package requests

// UpdateConversationRequest pairs the route id with the new label.
type UpdateConversationRequest struct {
	ConversationIdRequest
	ConversationNameRequest
}

func (r *UpdateConversationRequest) ValidateAndSanitize() error {
	if err := r.ConversationIdRequest.ValidateAndSanitize(); err != nil {
		return err
	}
	return r.ConversationNameRequest.ValidateAndSanitize()
}
