package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_db/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenApplyingSqliteDsnDefaults(t *testing.T) {
	t.Parallel()

	t.Run("Given a bare file path", func(t *testing.T) {
		t.Parallel()

		t.Run("Should append all default pragmas", func(t *testing.T) {
			t.Parallel()

			result := utils.ApplySqliteDsnDefaults("/data/agent.db")
			assert.Equal(
				t,
				"/data/agent.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)",
				result,
			)
		})
	})

	t.Run("Given a DSN with a non-pragma query string", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep the query and append all default pragmas", func(t *testing.T) {
			t.Parallel()

			result := utils.ApplySqliteDsnDefaults("file:agent.db?cache=shared")
			assert.Equal(
				t,
				"file:agent.db?cache=shared&_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)",
				result,
			)
		})
	})

	t.Run("Given a DSN that overrides one default pragma", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep the override and append only the missing defaults", func(t *testing.T) {
			t.Parallel()

			result := utils.ApplySqliteDsnDefaults("/data/agent.db?_pragma=busy_timeout(500)")
			assert.Equal(
				t,
				"/data/agent.db?_pragma=busy_timeout(500)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)",
				result,
			)
		})
	})

	t.Run("Given a DSN that sets all default pragmas", func(t *testing.T) {
		t.Parallel()

		t.Run("Should leave the DSN untouched", func(t *testing.T) {
			t.Parallel()

			dsn := "/data/agent.db?_pragma=foreign_keys(0)&_pragma=busy_timeout(500)&_pragma=journal_mode(DELETE)"
			assert.Equal(t, dsn, utils.ApplySqliteDsnDefaults(dsn))
		})
	})

	t.Run("Given a DSN with an empty _pragma parameter", func(t *testing.T) {
		t.Parallel()

		t.Run("Should still append all default pragmas", func(t *testing.T) {
			t.Parallel()

			result := utils.ApplySqliteDsnDefaults("/data/agent.db?_pragma=")
			assert.Equal(
				t,
				"/data/agent.db?_pragma=&_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)",
				result,
			)
		})
	})

	t.Run("Given a pragma using the name=value form", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recognize the override and skip that default", func(t *testing.T) {
			t.Parallel()

			result := utils.ApplySqliteDsnDefaults("/data/agent.db?_pragma=journal_mode%3DDELETE")
			assert.Equal(
				t,
				"/data/agent.db?_pragma=journal_mode%3DDELETE&_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)",
				result,
			)
		})
	})
}
