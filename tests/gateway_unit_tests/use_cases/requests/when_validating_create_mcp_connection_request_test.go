package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingCreateMcpConnectionRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given name and url are set without auth headers", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{
			Name: "search",
			Url:  "https://mcp.example.com",
		}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given an auth header name + value are both set", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{
			Name:            "search",
			Url:             "https://mcp.example.com",
			AuthHeaderName:  "Authorization",
			AuthHeaderValue: "Bearer xyz",
		}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given the name is whitespace", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{Name: "  ", Url: "https://mcp.example.com"}

		t.Run("Should reject with a validation error on name", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "name")
		})
	})

	t.Run("Given the url is empty", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{Name: "search", Url: ""}

		t.Run("Should reject with a validation error on url", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "url")
		})
	})

	t.Run("Given an auth header name is set but value is empty", func(t *testing.T) {
		t.Parallel()
		req := &requests.CreateMcpConnectionRequest{
			Name:            "search",
			Url:             "https://mcp.example.com",
			AuthHeaderName:  "Authorization",
			AuthHeaderValue: "  ",
		}

		t.Run("Should reject with a validation error on authHeaderValue", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "authHeaderValue")
		})
	})
}
