package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingMCPConnectionRequest(t *testing.T) {
	t.Parallel()

	generateMCPConnectionRequest := func() *requests.MCPConnectionRequest {
		return &requests.MCPConnectionRequest{
			Name:                "  my-tools  ",
			Url:                 "  https://my.mcp.com/v1  ",
			AllowedToolPatterns: []string{"  search_issues  "},
		}
	}

	t.Run("Given a blank name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := generateMCPConnectionRequest()
			req.Name = "   "

			err := req.ValidateAndSanitize()

			assert.ErrorContains(t, err, "name")
		})
	})

	t.Run("Given a name longer than 15 characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := generateMCPConnectionRequest()
			req.Name = strings.Repeat("a", requests.MaxMCPConnectionNameLength+1)

			err := req.ValidateAndSanitize()

			assert.ErrorContains(t, err, "name")
		})
	})

	t.Run("Given the name contains an underscore", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := generateMCPConnectionRequest()
			req.Name = "my_tools"

			err := req.ValidateAndSanitize()

			assert.ErrorContains(t, err, "name")
		})
	})

	t.Run("Given the name does not start with a letter", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"1tool", "-tool", "9"} {
				req := generateMCPConnectionRequest()
				req.Name = name

				err := req.ValidateAndSanitize()

				assert.ErrorContains(t, err, "name")
			}
		})
	})

	t.Run("Given the name contains disallowed characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"my tools", "my.tools", "wéather"} {
				req := generateMCPConnectionRequest()
				req.Name = name

				err := req.ValidateAndSanitize()

				assert.ErrorContains(t, err, "name")
			}
		})
	})

	t.Run("Given the name starts with mcp", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError regardless of case", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"mcptools", "MCP-tools", "Mcpthing"} {
				req := generateMCPConnectionRequest()
				req.Name = name

				err := req.ValidateAndSanitize()

				assert.ErrorContains(t, err, "name")
			}
		})
	})

	t.Run("Given the name is the reserved self name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError regardless of case", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"self", "Self", "SELF"} {
				req := generateMCPConnectionRequest()
				req.Name = name

				err := req.ValidateAndSanitize()

				assert.ErrorContains(t, err, "name")
			}
		})
	})

	t.Run("Given a blank url", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := generateMCPConnectionRequest()
			req.Url = "   "

			err := req.ValidateAndSanitize()

			assert.ErrorContains(t, err, "url")
		})
	})

	t.Run("Given no allowed tool pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := generateMCPConnectionRequest()
			req.AllowedToolPatterns = nil

			err := req.ValidateAndSanitize()

			assert.ErrorContains(t, err, "allowedToolPatterns")
		})
	})

	t.Run("Given an allowed tool pattern is blank", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := generateMCPConnectionRequest()
			req.AllowedToolPatterns = []string{"   "}

			err := req.ValidateAndSanitize()

			assert.ErrorContains(t, err, "allowedToolPatterns")
		})
	})

	t.Run("Given an allowed tool pattern is not a valid regex", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := generateMCPConnectionRequest()
			req.AllowedToolPatterns = []string{"search_[issues"}

			err := req.ValidateAndSanitize()

			assert.ErrorContains(t, err, "allowedToolPatterns")
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sanitize the fields", func(t *testing.T) {
			t.Parallel()

			req := generateMCPConnectionRequest()
			err := req.ValidateAndSanitize()

			assert.NoError(t, err)
			assert.Equal(t, "my-tools", req.Name)
			assert.Equal(t, "https://my.mcp.com/v1", req.Url)
			assert.Equal(t, []string{"search_issues"}, req.AllowedToolPatterns)
		})
	})
}
