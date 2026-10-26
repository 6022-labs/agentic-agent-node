package seeds

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
)

// SeedAgent inserts a parent agents row directly; AgentRepository.Upsert collapses to one row and cannot seed several.
func SeedAgent(dbClient *runtime_db.DatabaseClient, agentId uuid.UUID) {
	agent := model_builders.NewAgentBuilder().WithId(agentId).Build()
	if err := dbClient.GetDB(context.Background()).Create(agent).Error; err != nil {
		panic(err)
	}
}
