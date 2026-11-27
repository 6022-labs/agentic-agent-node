package requests

import (
	"net/url"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// CallAgentRequest sends one message to the agent whose base url is Url.
type CallAgentRequest struct {
	Url              string `json:"url"`
	Message          string `json:"message"`
	ContextId        string `json:"contextId,omitempty"`
	PaymentSignature string `json:"paymentSignature,omitempty"`
}

func (r *CallAgentRequest) ValidateAndSanitize() error {
	r.Url = strings.TrimRight(strings.TrimSpace(r.Url), "/")
	r.Message = strings.TrimSpace(r.Message)
	r.ContextId = strings.TrimSpace(r.ContextId)
	r.PaymentSignature = strings.TrimSpace(r.PaymentSignature)

	if len(r.Url) == 0 {
		return errors.NewValidationError("url", "cannot be empty")
	}
	parsed, err := url.Parse(r.Url)
	if err != nil {
		return errors.NewValidationError("url", "must be a valid url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.NewValidationError("url", "must use http or https")
	}
	if len(parsed.Hostname()) == 0 {
		return errors.NewValidationError("url", "must name a host")
	}
	if parsed.User != nil {
		return errors.NewValidationError("url", "cannot carry credentials")
	}
	// Paths are appended to the base url, so a query or fragment would swallow them.
	if strings.ContainsAny(r.Url, "?#") {
		return errors.NewValidationError("url", "cannot carry a query or fragment")
	}
	if len(r.Message) == 0 {
		return errors.NewValidationError("message", "cannot be empty")
	}
	return nil
}
