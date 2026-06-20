package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentPolicyRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentPolicyRepository(client *runtime_db.DatabaseClient) *PaymentPolicyRepository {
	return &PaymentPolicyRepository{client: client}
}

func (r *PaymentPolicyRepository) GetByAgentId(ctx context.Context, agentId uuid.UUID) (*models.PaymentPolicy, error) {
	var entity models.PaymentPolicy
	err := r.client.GetDB().WithContext(ctx).First(&entity, "agent_id = ?", agentId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *PaymentPolicyRepository) Create(ctx context.Context, policy *models.PaymentPolicy) (*models.PaymentPolicy, error) {
	if err := r.client.GetDB().WithContext(ctx).Create(policy).Error; err != nil {
		return nil, err
	}
	return policy, nil
}

func (r *PaymentPolicyRepository) DeleteByAgentId(ctx context.Context, agentId uuid.UUID) error {
	return r.client.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var policy models.PaymentPolicy
		err := tx.First(&policy, "agent_id = ?", agentId).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := tx.Where("policy_id = ?", policy.Id).Delete(&models.PaymentRule{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", policy.Id).Delete(&models.PaymentPolicy{}).Error
	})
}
