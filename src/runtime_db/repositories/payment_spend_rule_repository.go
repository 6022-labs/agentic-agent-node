package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentSpendRuleRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentSpendRuleRepository(client *runtime_db.DatabaseClient) *PaymentSpendRuleRepository {
	return &PaymentSpendRuleRepository{client: client}
}

func (r *PaymentSpendRuleRepository) ListByPolicyId(
	ctx context.Context,
	policyId uuid.UUID,
) ([]models.PaymentSpendRule, error) {
	rules := []models.PaymentSpendRule{}
	err := r.client.GetDB(ctx).Where("policy_id = ?", policyId).Order("position ASC").Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *PaymentSpendRuleRepository) GetById(ctx context.Context, id uuid.UUID) (*models.PaymentSpendRule, error) {
	var entity models.PaymentSpendRule
	err := r.client.GetDB(ctx).First(&entity, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *PaymentSpendRuleRepository) Create(
	ctx context.Context,
	rule *models.PaymentSpendRule,
) (*models.PaymentSpendRule, error) {
	err := r.client.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		var maxPosition *int
		if err := tx.Model(&models.PaymentSpendRule{}).
			Where("policy_id = ?", rule.PolicyId).
			Select("MAX(position)").
			Scan(&maxPosition).
			Error; err != nil {
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

func (r *PaymentSpendRuleRepository) Update(
	ctx context.Context,
	rule *models.PaymentSpendRule,
) (*models.PaymentSpendRule, error) {
	if err := r.client.GetDB(ctx).Omit("Policy").Save(rule).Error; err != nil {
		return nil, err
	}
	return rule, nil
}

func (r *PaymentSpendRuleRepository) DeleteById(ctx context.Context, id uuid.UUID) error {
	return r.client.GetDB(ctx).Delete(&models.PaymentSpendRule{}, "id = ?", id).Error
}
