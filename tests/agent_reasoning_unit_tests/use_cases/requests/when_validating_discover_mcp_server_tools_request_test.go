package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingDiscoverMCPServerToolsRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given neither a url nor a connection id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on url", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "url")
		})
	})

	t.Run("Given a blank url without a connection id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{Url: lo.ToPtr("   ")}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "url")
		})
	})

	t.Run("Given a blank header name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{
				Url:                  lo.ToPtr("https://my.mcp.com/v1"),
				MCPHeaderAuthRequest: requests.MCPHeaderAuthRequest{HeaderAuth: map[string]string{"  ": "value"}},
			}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "headerAuth")
		})
	})

	t.Run("Given a url only", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{Url: lo.ToPtr("  https://my.mcp.com/v1  ")}

			assert.NoError(t, req.ValidateAndSanitize())
		})

		t.Run("Should trim the url", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{Url: lo.ToPtr("  https://my.mcp.com/v1  ")}

			_ = req.ValidateAndSanitize()

			assert.Equal(t, "https://my.mcp.com/v1", lo.FromPtr(req.Url))
		})
	})

	t.Run("Given a connection id only", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{ConnectionId: lo.ToPtr(uuid.New())}

			assert.NoError(t, req.ValidateAndSanitize())
		})

		t.Run("Should leave the url unset", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{ConnectionId: lo.ToPtr(uuid.New())}

			_ = req.ValidateAndSanitize()

			assert.Nil(t, req.Url)
		})
	})

	t.Run("Given a blank header value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{
				ConnectionId:         lo.ToPtr(uuid.New()),
				MCPHeaderAuthRequest: requests.MCPHeaderAuthRequest{HeaderAuth: map[string]string{"Authorization": ""}},
			}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
