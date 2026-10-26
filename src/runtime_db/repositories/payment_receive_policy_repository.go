package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentReceivePolicyRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentReceivePolicyRepository(client *runtime_db.DatabaseClient) *PaymentReceivePolicyRepository {
	return &PaymentReceivePolicyRepository{client: client}
}

func (r *PaymentReceivePolicyRepository) GetByAgentId(
	ctx context.Context,
	agentId uuid.UUID,
) (*models.PaymentReceivePolicy, error) {
	var entity models.PaymentReceivePolicy
	err := r.client.GetDB(ctx).First(&entity, "agent_id = ?", agentId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *PaymentReceivePolicyRepository) Create(
	ctx context.Context,
	policy *models.PaymentReceivePolicy,
) (*models.PaymentReceivePolicy, error) {
	if err := r.client.GetDB(ctx).Create(policy).Error; err != nil {
		return nil, err
	}
	return policy, nil
}

func (r *PaymentReceivePolicyRepository) DeleteByAgentId(ctx context.Context, agentId uuid.UUID) error {
	return r.client.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		var policy models.PaymentReceivePolicy
		err := tx.First(&policy, "agent_id = ?", agentId).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := tx.Where("policy_id = ?", policy.Id).Delete(&models.PaymentReceiveRule{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", policy.Id).Delete(&models.PaymentReceivePolicy{}).Error
	})
}
