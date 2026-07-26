package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestWhenFilteringToolsByPatterns(t *testing.T) {
	t.Parallel()

	tools := []*models.Tool{
		{Name: "search_issues"},
		{Name: "create_issue"},
		{Name: "delete_repo"},
	}
	names := func(filtered []*models.Tool) []string {
		return lo.Map(filtered, func(tool *models.Tool, _ int) string { return tool.Name })
	}

	t.Run("Given no pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep every tool", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tools, utils.FilterToolsByPatterns(tools, nil))
		})
	})

	t.Run("Given exact tool names", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep only those tools", func(t *testing.T) {
			t.Parallel()

			filtered := utils.FilterToolsByPatterns(tools, []string{"search_issues", "create_issue"})

			assert.Equal(t, []string{"search_issues", "create_issue"}, names(filtered))
		})
	})

	t.Run("Given a pattern that only matches part of a name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not keep the tool, since patterns are full matches", func(t *testing.T) {
			t.Parallel()

			assert.Empty(t, utils.FilterToolsByPatterns(tools, []string{"issue"}))
		})
	})

	t.Run("Given a wildcard pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep every tool", func(t *testing.T) {
			t.Parallel()

			filtered := utils.FilterToolsByPatterns(tools, []string{".*"})

			assert.Equal(t, []string{"search_issues", "create_issue", "delete_repo"}, names(filtered))
		})
	})

	t.Run("Given a prefix pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep the matching tools only", func(t *testing.T) {
			t.Parallel()

			filtered := utils.FilterToolsByPatterns(tools, []string{"search_.*"})

			assert.Equal(t, []string{"search_issues"}, names(filtered))
		})
	})

	t.Run("Given an invalid pattern alongside a valid one", func(t *testing.T) {
		t.Parallel()

		t.Run("Should ignore the invalid pattern and keep what the valid one matches", func(t *testing.T) {
			t.Parallel()

			filtered := utils.FilterToolsByPatterns(tools, []string{"create_[issue", "create_issue"})

			assert.Equal(t, []string{"create_issue"}, names(filtered))
		})
	})
}
