package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// JsonContentType is the content type sent on every bodied runtime request.
const JsonContentType = "application/json"

// RuntimeStatusError carries a non-2xx runtime response so callers can preserve
// the semantic status (404/409/400). Error() keeps the legacy string form.
type RuntimeStatusError struct {
	StatusCode int
	Code       string
	Field      string
	Message    string
	Body       string
}

func (e *RuntimeStatusError) Error() string {
	return fmt.Sprintf("runtime returned %d: %s", e.StatusCode, e.Body)
}

// DecodeRuntimeResponse reproduces the legacy transport's handling for a generated
// oapi response: notFoundErr (when non-nil) is returned on 404, any other non-2xx
// yields a *RuntimeStatusError parsed from the body, and a 2xx body is decoded
// into out (skipped when out is nil).
func DecodeRuntimeResponse(httpResponse *http.Response, body []byte, out any, notFoundErr error) error {
	if notFoundErr != nil && httpResponse.StatusCode == http.StatusNotFound {
		return notFoundErr
	}
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		statusErr := &RuntimeStatusError{StatusCode: httpResponse.StatusCode, Body: string(body)}
		var parsed struct {
			Code    string `json:"code"`
			Field   string `json:"field"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &parsed) == nil {
			statusErr.Code = parsed.Code
			statusErr.Field = parsed.Field
			statusErr.Message = parsed.Message
		}
		return statusErr
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(body, out)
}

// MapRuntimeError maps the runtime's HTTP-status error to a domain error,
// preserving 400/404/409 and collapsing the rest to 502.
func MapRuntimeError(err error) error {
	var statusErr *RuntimeStatusError
	if errors.As(err, &statusErr) {
		switch statusErr.StatusCode {
		case http.StatusBadRequest:
			return apperrors.NewValidationError(statusErr.Field, statusErr.Message)
		case http.StatusNotFound:
			return apperrors.NewNotFoundError(statusErr.Code, statusErr.Message)
		case http.StatusConflict:
			return apperrors.NewConflictError(statusErr.Code, statusErr.Message)
		}
	}
	return apperrors.NewUnavailableError("runtime_unavailable", err.Error())
}
