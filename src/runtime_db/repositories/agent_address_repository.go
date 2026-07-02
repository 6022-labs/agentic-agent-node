package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AgentAddressRepository struct {
	client *runtime_db.DatabaseClient
}

func NewAgentAddressRepository(client *runtime_db.DatabaseClient) *AgentAddressRepository {
	return &AgentAddressRepository{client: client}
}

func (r *AgentAddressRepository) Upsert(ctx context.Context, address *models.AgentAddress) error {
	return r.client.GetDB().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "agent_id"}, {Name: "type"}},
			DoUpdates: clause.AssignmentColumns([]string{"address", "updated_at"}),
		}).
		Create(address).Error
}

func (r *AgentAddressRepository) GetByAgentIdAndType(ctx context.Context, agentId uuid.UUID, addressType enums.AddressType) (*models.AgentAddress, error) {
	var entity models.AgentAddress
	err := r.client.GetDB().WithContext(ctx).First(&entity, "agent_id = ? AND type = ?", agentId, addressType).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
