package repositories_test

import (
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
)

// seedAgent inserts a parent agents row so child rows satisfy the agent_id FK.
// Uses a direct insert rather than AgentRepository.Upsert, which collapses to
// the single existing row and so cannot seed multiple distinct agents.
func seedAgent(dbClient *runtime_db.DatabaseClient, agentId uuid.UUID) {
	agent := model_builders.NewAgentBuilder().WithId(agentId).Build()
	if err := dbClient.GetDB().Create(agent).Error; err != nil {
		panic(err)
	}
}
