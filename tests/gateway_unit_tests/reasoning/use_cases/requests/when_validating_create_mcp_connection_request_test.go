package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingCreateMcpConnectionRequest(t *testing.T) {
	t.Parallel()

	generateCreateMcpConnectionRequest := func() *requests.CreateMcpConnectionRequest {
		return &requests.CreateMcpConnectionRequest{
			McpConnectionRequest: requests.McpConnectionRequest{
				Name:                "search",
				Url:                 "https://mcp.example.com",
				AllowedToolPatterns: []string{"search_issues"},
			},
			McpHeaderAuthRequest: requests.McpHeaderAuthRequest{
				HeaderAuth: map[string]string{"Authorization": "Bearer xyz"},
			},
		}
	}

	t.Run("Given a blank name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on name", func(t *testing.T) {
			t.Parallel()

			req := generateCreateMcpConnectionRequest()
			req.Name = "  "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a blank header name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on headerAuth", func(t *testing.T) {
			t.Parallel()

			req := generateCreateMcpConnectionRequest()
			req.HeaderAuth = map[string]string{"  ": "Bearer xyz"}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "headerAuth")
		})
	})

	t.Run("Given a blank header value", func(t *testing.T) {
		t.Parallel()

		t.Run(
			"Should reject with a validation error on headerAuth, since nothing is stored to fall back to",
			func(t *testing.T) {
				t.Parallel()

				req := generateCreateMcpConnectionRequest()
				req.HeaderAuth = map[string]string{"Authorization": "  "}

				assert.ErrorContains(t, req.ValidateAndSanitize(), "headerAuth")
			},
		)
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, generateCreateMcpConnectionRequest().ValidateAndSanitize())
		})
	})
}
