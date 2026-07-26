package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingCreateMcpConnectionRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given name and url without headers", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{
			Name:                "search",
			Url:                 "https://mcp.example.com",
			AllowedToolPatterns: []string{"search_issues"},
		}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given a header name and value", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{
			Name:                "search",
			Url:                 "https://mcp.example.com",
			AllowedToolPatterns: []string{"search_issues"},
			HeaderAuth:          map[string]string{"Authorization": "Bearer xyz"},
		}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given the name is whitespace", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{Name: "  ", Url: "https://mcp.example.com", AllowedToolPatterns: []string{"search_issues"}}

		t.Run("Should reject with a validation error on name", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "name")
		})
	})

	t.Run("Given the url is empty", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{Name: "search", Url: "", AllowedToolPatterns: []string{"search_issues"}}

		t.Run("Should reject with a validation error on url", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "url")
		})
	})

	t.Run("Given a header value is empty", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{
			Name:                "search",
			Url:                 "https://mcp.example.com",
			AllowedToolPatterns: []string{"search_issues"},
			HeaderAuth:          map[string]string{"Authorization": "  "},
		}

		t.Run("Should reject with a validation error on headerAuth", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "headerAuth")
		})
	})

	t.Run("Given no allowed tool", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{Name: "search", Url: "https://mcp.example.com"}

		t.Run("Should reject with a validation error on allowedToolPatterns", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "allowedToolPatterns")
		})
	})

	t.Run("Given a header name is empty", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{
			Name:                "search",
			Url:                 "https://mcp.example.com",
			AllowedToolPatterns: []string{"search_issues"},
			HeaderAuth:          map[string]string{"  ": "Bearer xyz"},
		}

		t.Run("Should reject with a validation error on headerAuth", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "headerAuth")
		})
	})
}
