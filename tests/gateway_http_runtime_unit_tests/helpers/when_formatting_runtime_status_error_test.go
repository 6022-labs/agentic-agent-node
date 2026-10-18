package helpers_test

import (
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenFormattingRuntimeStatusError(t *testing.T) {
	t.Parallel()

	t.Run("Given a RuntimeStatusError with a status and body", func(t *testing.T) {
		t.Parallel()

		t.Run("Should format the legacy string", func(t *testing.T) {
			t.Parallel()

			statusErr := &helpers.RuntimeStatusError{StatusCode: http.StatusConflict, Body: `{"code":"not_minted"}`}

			assert.Equal(t, `runtime returned 409: {"code":"not_minted"}`, statusErr.Error())
		})
	})
}
