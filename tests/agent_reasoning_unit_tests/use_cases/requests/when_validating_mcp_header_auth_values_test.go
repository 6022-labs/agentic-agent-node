package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingMCPHeaderAuthValues(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank header value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the headerAuth field", func(t *testing.T) {
			t.Parallel()

			req := &requests.MCPHeaderAuthRequest{HeaderAuth: map[string]string{"Authorization": "  "}}

			assert.ErrorContains(t, req.ValidateValues(), "headerAuth")
		})
	})

	t.Run("Given no header auth", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := &requests.MCPHeaderAuthRequest{}

			assert.NoError(t, req.ValidateValues())
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := &requests.MCPHeaderAuthRequest{HeaderAuth: map[string]string{"Authorization": "Bearer secret"}}

			assert.NoError(t, req.ValidateValues())
		})
	})
}
