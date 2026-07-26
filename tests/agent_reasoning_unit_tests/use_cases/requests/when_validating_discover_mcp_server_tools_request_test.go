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

	t.Run("Given a url only", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation and trim the url", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{Url: lo.ToPtr("  https://my.mcp.com/v1  ")}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "https://my.mcp.com/v1", lo.FromPtr(req.Url))
		})
	})

	t.Run("Given a connection id only", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation, since the stored connection supplies the url", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{ConnectionId: lo.ToPtr(uuid.New())}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Nil(t, req.Url)
		})
	})

	t.Run("Given neither a url nor a connection id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on url", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "url")
		})
	})

	t.Run("Given a blank url without a connection id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error, since trimming leaves nothing", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{Url: lo.ToPtr("   ")}

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a header auth entry has an empty name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{
				Url:        lo.ToPtr("https://my.mcp.com/v1"),
				HeaderAuth: map[string]string{"  ": "value"},
			}

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a header auth entry has an empty value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation, since blank falls back to the stored value", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMCPServerToolsRequest{
				ConnectionId: lo.ToPtr(uuid.New()),
				HeaderAuth:   map[string]string{"Authorization": ""},
			}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
