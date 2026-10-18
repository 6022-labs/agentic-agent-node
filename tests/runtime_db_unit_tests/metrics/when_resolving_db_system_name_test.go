package metrics_test

import (
	"testing"

	knowledge_enums "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/metrics"
	"github.com/stretchr/testify/assert"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

func TestWhenResolvingDbSystemName(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		driver   knowledge_enums.DatabaseDriver
		expected string
	}{
		"mysql":    {knowledge_enums.DatabaseDriverMySQL, semconv.DBSystemNameMySQL.Value.AsString()},
		"mariadb":  {knowledge_enums.DatabaseDriverMariaDB, semconv.DBSystemNameMariaDB.Value.AsString()},
		"sqlite":   {knowledge_enums.DatabaseDriverSqlite, semconv.DBSystemNameSQLite.Value.AsString()},
		"postgres": {knowledge_enums.DatabaseDriverPostgres, semconv.DBSystemNamePostgreSQL.Value.AsString()},
		"unmapped": {knowledge_enums.DatabaseDriver("unknown"), semconv.DBSystemNamePostgreSQL.Value.AsString()},
	}

	for name, c := range cases {
		t.Run("Given the "+name+" driver", func(t *testing.T) {
			t.Parallel()

			t.Run("Should stamp the matching db.system.name", func(t *testing.T) {
				t.Parallel()

				sut := metrics.NewDbQueryMetrics(c.driver)

				assert.Equal(t, c.expected, sut.DbSystemName().Value.AsString())
			})
		})
	}
}
