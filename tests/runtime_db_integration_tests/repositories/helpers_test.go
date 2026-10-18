package repositories_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// seedAgent inserts a parent agents row directly; AgentRepository.Upsert collapses to one row and cannot seed several.
func seedAgent(dbClient *runtime_db.DatabaseClient, agentId uuid.UUID) {
	agent := model_builders.NewAgentBuilder().WithId(agentId).Build()
	if err := dbClient.GetDB().Create(agent).Error; err != nil {
		panic(err)
	}
}

// assertUniqueViolation matches the driver-specific duplicate-key wording of postgres, mysql/mariadb and sqlite.
func assertUniqueViolation(t *testing.T, err error) {
	t.Helper()
	if assert.Error(t, err) {
		msg := strings.ToLower(err.Error())
		assert.True(
			t,
			strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique"),
			"not a unique violation: %v",
			err,
		)
	}
}
