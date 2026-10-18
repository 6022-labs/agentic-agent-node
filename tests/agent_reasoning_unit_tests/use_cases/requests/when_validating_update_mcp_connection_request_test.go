package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingUpdateMCPConnectionRequest(t *testing.T) {
	t.Parallel()

	generateUpdateMCPConnectionRequest := func() *requests.UpdateMCPConnectionRequest {
		return &requests.UpdateMCPConnectionRequest{
			Id: uuid.New(),
			MCPConnectionRequest: requests.MCPConnectionRequest{
				Name:                "my-tools",
				Url:                 "https://my.mcp.com/v1",
				AllowedToolPatterns: []string{"search_issues"},
			},
			MCPHeaderAuthRequest: requests.MCPHeaderAuthRequest{
				HeaderAuth: map[string]string{"Authorization": "Bearer secret"},
			},
		}
	}

	t.Run("Given a blank name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the name field", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMCPConnectionRequest()
			req.Name = "   "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a blank header name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the headerAuth field", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMCPConnectionRequest()
			req.HeaderAuth = map[string]string{"  ": "value"}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "headerAuth")
		})
	})

	t.Run("Given a blank header value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMCPConnectionRequest()
			req.HeaderAuth = map[string]string{"Authorization": "  "}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, generateUpdateMCPConnectionRequest().ValidateAndSanitize())
		})
	})
}
