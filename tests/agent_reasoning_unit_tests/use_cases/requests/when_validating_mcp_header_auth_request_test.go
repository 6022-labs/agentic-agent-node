package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingMCPHeaderAuthRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank header name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the headerAuth field", func(t *testing.T) {
			t.Parallel()

			req := &requests.MCPHeaderAuthRequest{HeaderAuth: map[string]string{"  ": "value"}}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "headerAuth")
		})
	})

	t.Run("Given a blank header value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass sanitizing", func(t *testing.T) {
			t.Parallel()

			req := &requests.MCPHeaderAuthRequest{HeaderAuth: map[string]string{"Authorization": "  "}}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given no header auth", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass sanitizing", func(t *testing.T) {
			t.Parallel()

			req := &requests.MCPHeaderAuthRequest{}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass sanitizing", func(t *testing.T) {
			t.Parallel()

			req := &requests.MCPHeaderAuthRequest{HeaderAuth: map[string]string{"Authorization": "Bearer secret"}}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
