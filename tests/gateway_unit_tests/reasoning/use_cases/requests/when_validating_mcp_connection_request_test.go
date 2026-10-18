package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingMcpConnectionRequest(t *testing.T) {
	t.Parallel()

	generateMcpConnectionRequest := func() *requests.McpConnectionRequest {
		return &requests.McpConnectionRequest{
			Name:                "search",
			Url:                 "https://mcp.example.com",
			AllowedToolPatterns: []string{"search_issues"},
		}
	}

	t.Run("Given a blank name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on name", func(t *testing.T) {
			t.Parallel()

			req := generateMcpConnectionRequest()
			req.Name = "  "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a blank url", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on url", func(t *testing.T) {
			t.Parallel()

			req := generateMcpConnectionRequest()
			req.Url = ""

			assert.ErrorContains(t, req.ValidateAndSanitize(), "url")
		})
	})

	t.Run("Given no allowed tool", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on allowedToolPatterns", func(t *testing.T) {
			t.Parallel()

			req := generateMcpConnectionRequest()
			req.AllowedToolPatterns = nil

			assert.ErrorContains(t, req.ValidateAndSanitize(), "allowedToolPatterns")
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, generateMcpConnectionRequest().ValidateAndSanitize())
		})
	})
}
