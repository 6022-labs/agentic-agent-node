package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingDiscoverMcpServerToolsRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a url only", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation and trim the url", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMcpServerToolsRequest{Url: lo.ToPtr("  https://mcp.example.com  ")}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "https://mcp.example.com", lo.FromPtr(req.Url))
		})
	})

	t.Run("Given a connection id only", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation, since the stored connection supplies the url", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMcpServerToolsRequest{ConnectionId: lo.ToPtr(uuid.New())}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given neither a url nor a connection id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on url", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMcpServerToolsRequest{}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "url")
		})
	})

	t.Run("Given a header with an empty name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on headerAuth", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMcpServerToolsRequest{
				Url:        lo.ToPtr("https://mcp.example.com"),
				HeaderAuth: map[string]string{"  ": "Bearer xyz"},
			}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "headerAuth")
		})
	})

	t.Run("Given headers sent without values", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation, since blank falls back to the stored values", func(t *testing.T) {
			t.Parallel()

			req := &requests.DiscoverMcpServerToolsRequest{
				ConnectionId: lo.ToPtr(uuid.New()),
				HeaderAuth:   map[string]string{"Authorization": "", "X-Api-Key": ""},
			}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
