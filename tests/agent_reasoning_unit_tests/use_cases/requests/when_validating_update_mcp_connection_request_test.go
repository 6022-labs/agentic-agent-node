package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingUpdateMCPConnectionRequest(t *testing.T) {
	t.Parallel()

	generateUpdateMCPConnectionRequest := func() *requests.UpdateMCPConnectionRequest {
		return &requests.UpdateMCPConnectionRequest{
			Name:       "  my-tools  ",
			Url:        "  https://my.mcp.com/v1  ",
			HeaderAuth: map[string]string{"Authorization": "Bearer secret"},
		}
	}

	t.Run("Given the request is valid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sanitize the fields", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMCPConnectionRequest()
			err := req.ValidateAndSanitize()

			assert.NoError(t, err)
			assert.Equal(t, "my-tools", req.Name)
			assert.Equal(t, "https://my.mcp.com/v1", req.Url)
		})
	})

	t.Run("Given the name is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMCPConnectionRequest()
			req.Name = "   "

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the url is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMCPConnectionRequest()
			req.Url = "   "

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a header auth entry has an empty value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateUpdateMCPConnectionRequest()
			req.HeaderAuth = map[string]string{"Authorization": "  "}

			assert.Error(t, req.ValidateAndSanitize())
		})
	})
}
