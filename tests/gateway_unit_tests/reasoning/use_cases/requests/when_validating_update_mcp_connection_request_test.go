package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingUpdateMcpConnectionRequest(t *testing.T) {
	t.Parallel()

	generateUpdateMcpConnectionRequest := func() *requests.UpdateMcpConnectionRequest {
		return &requests.UpdateMcpConnectionRequest{
			Id: uuid.New(),
			McpConnectionRequest: requests.McpConnectionRequest{
				Name:                "search",
				Url:                 "https://mcp.example.com",
				AllowedToolPatterns: []string{"search_issues"},
			},
			McpHeaderAuthRequest: requests.McpHeaderAuthRequest{
				HeaderAuth: map[string]string{"Authorization": ""},
			},
		}
	}

	t.Run("Given a blank name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on name", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMcpConnectionRequest()
			req.Name = "  "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a blank header name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on headerAuth", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMcpConnectionRequest()
			req.HeaderAuth = map[string]string{"  ": "Bearer xyz"}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "headerAuth")
		})
	})

	t.Run("Given a blank header value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation, since blank keeps the stored value", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, generateUpdateMcpConnectionRequest().ValidateAndSanitize())
		})
	})

	t.Run("Given no header auth", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation, since dropping the header is allowed", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMcpConnectionRequest()
			req.HeaderAuth = nil

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
