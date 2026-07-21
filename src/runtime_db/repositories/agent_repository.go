package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"gorm.io/gorm"
)

type AgentRepository struct {
	client *runtime_db.DatabaseClient
}

func NewAgentRepository(client *runtime_db.DatabaseClient) *AgentRepository {
	return &AgentRepository{client: client}
}

func (r *AgentRepository) Get(ctx context.Context) (*models.Agent, error) {
	var agent models.Agent
	err := r.client.GetDB().WithContext(ctx).First(&agent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) Upsert(ctx context.Context, agent *models.Agent) (*models.Agent, error) {
	existing, err := r.Get(ctx)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		if err := r.client.GetDB().WithContext(ctx).Create(agent).Error; err != nil {
			return nil, err
		}
		return agent, nil
	}

	existing.ChainId = agent.ChainId
	existing.CollectionTokenId = agent.CollectionTokenId
	existing.CollectionAddress = agent.CollectionAddress
	existing.ProposalId = agent.ProposalId
	existing.Name = agent.Name
	existing.CloneOf = agent.CloneOf
	existing.EnsDomain = agent.EnsDomain
	existing.Creator = agent.Creator
	existing.Owner = agent.Owner
	existing.Images = agent.Images
	existing.Attributes = agent.Attributes
	if err := r.client.
		GetDB().
		WithContext(ctx).
		Save(existing).
		Error; err != nil {
		return nil, err
	}
	return existing, nil
}
