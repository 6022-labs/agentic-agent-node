package helpers_test

import (
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenFormattingRuntimeStatusError(t *testing.T) {
	t.Parallel()

	t.Run("Given a RuntimeStatusError carrying a code", func(t *testing.T) {
		t.Parallel()

		t.Run("Should name the code", func(t *testing.T) {
			t.Parallel()

			statusErr := &helpers.RuntimeStatusError{
				StatusCode: http.StatusConflict,
				Code:       "not_minted",
				Body:       `{"code":"not_minted","message":"host=db user=postgres"}`,
			}

			assert.Equal(t, "runtime returned 409: not_minted", statusErr.Error())
		})

		// Gateway use cases put this string into responses, and /responses is unauthenticated.
		t.Run("Should never carry the response body", func(t *testing.T) {
			t.Parallel()

			statusErr := &helpers.RuntimeStatusError{
				StatusCode: http.StatusInternalServerError,
				Code:       "inference_failed",
				Body:       `{"code":"inference_failed","message":"dial tcp 10.0.0.5:5432: connect: refused"}`,
			}

			assert.NotContains(t, statusErr.Error(), "10.0.0.5")
			assert.NotContains(t, statusErr.Error(), statusErr.Body)
		})
	})

	t.Run("Given a RuntimeStatusError with no parsable code", func(t *testing.T) {
		t.Parallel()

		t.Run("Should name only the status", func(t *testing.T) {
			t.Parallel()

			statusErr := &helpers.RuntimeStatusError{StatusCode: http.StatusBadGateway, Body: "<html>nginx</html>"}

			assert.Equal(t, "runtime returned 502", statusErr.Error())
		})
	})
}
