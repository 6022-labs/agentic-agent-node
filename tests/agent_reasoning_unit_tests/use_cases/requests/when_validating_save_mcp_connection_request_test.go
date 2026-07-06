package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSaveMCPConnectionRequest(t *testing.T) {
	t.Parallel()

	generateSaveMCPConnectionRequest := func() *requests.SaveMCPConnectionRequest {
		return &requests.SaveMCPConnectionRequest{
			Name:       "  my-tools  ",
			Url:        "  https://my.mcp.com/v1  ",
			HeaderAuth: map[string]string{"Authorization": "Bearer secret"},
		}
	}

	t.Run("Given the request is valid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sanitize the fields", func(t *testing.T) {
			t.Parallel()

			req := generateSaveMCPConnectionRequest()
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

			req := generateSaveMCPConnectionRequest()
			req.Name = "   "

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the name is the reserved self name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error regardless of case", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"self", "Self", "SELF"} {
				req := generateSaveMCPConnectionRequest()
				req.Name = name

				assert.Error(t, req.ValidateAndSanitize())
			}
		})
	})

	t.Run("Given the name contains an underscore", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveMCPConnectionRequest()
			req.Name = "my_tools"

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the name is too long", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveMCPConnectionRequest()
			req.Name = strings.Repeat("a", requests.MaxMCPConnectionNameLength+1)

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the name starts with mcp", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error regardless of case", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"mcptools", "MCP-tools", "Mcpthing"} {
				req := generateSaveMCPConnectionRequest()
				req.Name = name

				assert.Error(t, req.ValidateAndSanitize())
			}
		})
	})

	t.Run("Given the name does not start with a letter", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"1tool", "-tool", "9"} {
				req := generateSaveMCPConnectionRequest()
				req.Name = name

				assert.Error(t, req.ValidateAndSanitize())
			}
		})
	})

	t.Run("Given the name contains disallowed characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"my tools", "my.tools", "wéather"} {
				req := generateSaveMCPConnectionRequest()
				req.Name = name

				assert.Error(t, req.ValidateAndSanitize())
			}
		})
	})

	t.Run("Given the url is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveMCPConnectionRequest()
			req.Url = "   "

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a header auth entry has an empty name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveMCPConnectionRequest()
			req.HeaderAuth = map[string]string{"  ": "value"}

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a header auth entry has an empty value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveMCPConnectionRequest()
			req.HeaderAuth = map[string]string{"Authorization": "  "}

			assert.Error(t, req.ValidateAndSanitize())
		})
	})
}
