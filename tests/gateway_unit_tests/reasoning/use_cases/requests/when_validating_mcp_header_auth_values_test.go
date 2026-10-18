package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingMcpHeaderAuthValues(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank header value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on headerAuth", func(t *testing.T) {
			t.Parallel()

			req := &requests.McpHeaderAuthRequest{HeaderAuth: map[string]string{"Authorization": " "}}

			assert.ErrorContains(t, req.ValidateValues(), "headerAuth")
		})
	})

	t.Run("Given no header auth", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := &requests.McpHeaderAuthRequest{}

			assert.NoError(t, req.ValidateValues())
		})
	})

	t.Run("Given a header name and value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := &requests.McpHeaderAuthRequest{HeaderAuth: map[string]string{"Authorization": "Bearer xyz"}}

			assert.NoError(t, req.ValidateValues())
		})
	})
}
