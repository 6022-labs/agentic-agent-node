package repositories

import (
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"gorm.io/gorm"
)

type LLMConnectionRepository struct {
	client *runtime_db.DatabaseClient
}

func NewLLMConnectionRepository(client *runtime_db.DatabaseClient) *LLMConnectionRepository {
	return &LLMConnectionRepository{client: client}
}

func (r *LLMConnectionRepository) Get() (*models.LLMConnection, error) {
	var entity models.LLMConnection
	err := r.client.GetDB().First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Upsert keeps a single row: update in place if one exists, otherwise insert.
func (r *LLMConnectionRepository) Upsert(connection *models.LLMConnection) (*models.LLMConnection, error) {
	existing, err := r.Get()
	if err != nil {
		return nil, err
	}
	if existing == nil {
		if err := r.client.GetDB().Create(connection).Error; err != nil {
			return nil, err
		}
		return connection, nil
	}

	existing.Platform = connection.Platform
	existing.Model = connection.Model
	existing.BaseUrl = connection.BaseUrl
	existing.SetupStatus = connection.SetupStatus
	if err := r.client.GetDB().Save(existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}

func (r *LLMConnectionRepository) Delete() error {
	return r.client.GetDB().
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(&models.LLMConnection{}).Error
}
