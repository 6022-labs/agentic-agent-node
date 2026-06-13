package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"gorm.io/gorm"
)

type PaymentPolicyRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentPolicyRepository(client *runtime_db.DatabaseClient) *PaymentPolicyRepository {
	return &PaymentPolicyRepository{client: client}
}

func (r *PaymentPolicyRepository) Get(ctx context.Context) (*models.PaymentPolicy, error) {
	var entity models.PaymentPolicy
	err := r.client.GetDB().WithContext(ctx).First(&entity).Error
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

func (r *PaymentPolicyRepository) Delete(ctx context.Context) error {
	return r.client.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session := tx.Session(&gorm.Session{AllowGlobalUpdate: true})
		if err := session.Delete(&models.PaymentRule{}).Error; err != nil {
			return err
		}
		return session.Delete(&models.PaymentPolicy{}).Error
	})
}
