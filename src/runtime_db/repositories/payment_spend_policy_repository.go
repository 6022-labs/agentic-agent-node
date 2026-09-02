package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentSpendPolicyRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentSpendPolicyRepository(client *runtime_db.DatabaseClient) *PaymentSpendPolicyRepository {
	return &PaymentSpendPolicyRepository{client: client}
}

func (r *PaymentSpendPolicyRepository) GetByAgentId(
	ctx context.Context,
	agentId uuid.UUID,
) (*models.PaymentSpendPolicy, error) {
	var entity models.PaymentSpendPolicy
	err := r.client.GetDB(ctx).First(&entity, "agent_id = ?", agentId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *PaymentSpendPolicyRepository) Create(
	ctx context.Context,
	policy *models.PaymentSpendPolicy,
) (*models.PaymentSpendPolicy, error) {
	if err := r.client.GetDB(ctx).Create(policy).Error; err != nil {
		return nil, err
	}
	return policy, nil
}

func (r *PaymentSpendPolicyRepository) DeleteByAgentId(ctx context.Context, agentId uuid.UUID) error {
	return r.client.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		var policy models.PaymentSpendPolicy
		err := tx.First(&policy, "agent_id = ?", agentId).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := tx.Where("policy_id = ?", policy.Id).Delete(&models.PaymentSpendRule{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", policy.Id).Delete(&models.PaymentSpendPolicy{}).Error
	})
}
