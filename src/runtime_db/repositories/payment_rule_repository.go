package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRuleRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentRuleRepository(client *runtime_db.DatabaseClient) *PaymentRuleRepository {
	return &PaymentRuleRepository{client: client}
}

func (r *PaymentRuleRepository) ListByPolicyId(ctx context.Context, policyId uuid.UUID) ([]models.PaymentRule, error) {
	rules := []models.PaymentRule{}
	err := r.client.GetDB().WithContext(ctx).Where("policy_id = ?", policyId).Order("position ASC").Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *PaymentRuleRepository) GetById(ctx context.Context, id uuid.UUID) (*models.PaymentRule, error) {
	var entity models.PaymentRule
	err := r.client.GetDB().WithContext(ctx).First(&entity, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *PaymentRuleRepository) Create(ctx context.Context, rule *models.PaymentRule) (*models.PaymentRule, error) {
	err := r.client.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxPosition *int
		if err := tx.Model(&models.PaymentRule{}).Where("policy_id = ?", rule.PolicyId).Select("MAX(position)").Scan(&maxPosition).Error; err != nil {
			return err
		}
		if maxPosition != nil {
			rule.Position = *maxPosition + 1
		}
		return tx.Omit("Policy").Create(rule).Error
	})
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (r *PaymentRuleRepository) Update(ctx context.Context, rule *models.PaymentRule) (*models.PaymentRule, error) {
	if err := r.client.GetDB().WithContext(ctx).Omit("Policy").Save(rule).Error; err != nil {
		return nil, err
	}
	return rule, nil
}

func (r *PaymentRuleRepository) DeleteById(ctx context.Context, id uuid.UUID) error {
	return r.client.GetDB().WithContext(ctx).Delete(&models.PaymentRule{}, "id = ?", id).Error
}
