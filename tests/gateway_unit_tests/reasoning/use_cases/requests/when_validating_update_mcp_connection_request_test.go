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
			Id:                  uuid.New(),
			Name:                "search",
			Url:                 "https://mcp.example.com",
			AllowedToolPatterns: []string{"search_issues"},
			HeaderAuth:          map[string]string{"Authorization": ""},
		}
	}

	t.Run("Given a header without a value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation, since blank keeps the stored value", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, generateUpdateMcpConnectionRequest().ValidateAndSanitize())
		})
	})

	t.Run("Given a header name and value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMcpConnectionRequest()
			req.HeaderAuth = map[string]string{"Authorization": "Bearer xyz"}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given no header at all", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation, since dropping the header is allowed", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMcpConnectionRequest()
			req.HeaderAuth = nil

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a header with an empty name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on headerAuth", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMcpConnectionRequest()
			req.HeaderAuth = map[string]string{"  ": "Bearer xyz"}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "headerAuth")
		})
	})

	t.Run("Given the name is whitespace", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on name", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMcpConnectionRequest()
			req.Name = "  "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given the url is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on url", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMcpConnectionRequest()
			req.Url = ""

			assert.ErrorContains(t, req.ValidateAndSanitize(), "url")
		})
	})

	t.Run("Given no allowed tool", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on allowedToolPatterns", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMcpConnectionRequest()
			req.AllowedToolPatterns = nil

			assert.ErrorContains(t, req.ValidateAndSanitize(), "allowedToolPatterns")
		})
	})
}
