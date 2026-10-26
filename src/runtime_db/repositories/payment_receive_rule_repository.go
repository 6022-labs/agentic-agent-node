package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentReceiveRuleRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentReceiveRuleRepository(client *runtime_db.DatabaseClient) *PaymentReceiveRuleRepository {
	return &PaymentReceiveRuleRepository{client: client}
}

func (r *PaymentReceiveRuleRepository) ListByPolicyId(
	ctx context.Context,
	policyId uuid.UUID,
) ([]models.PaymentReceiveRule, error) {
	rules := []models.PaymentReceiveRule{}
	err := r.client.GetDB(ctx).Where("policy_id = ?", policyId).Order("position ASC").Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *PaymentReceiveRuleRepository) GetById(ctx context.Context, id uuid.UUID) (*models.PaymentReceiveRule, error) {
	var entity models.PaymentReceiveRule
	err := r.client.GetDB(ctx).First(&entity, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *PaymentReceiveRuleRepository) Create(
	ctx context.Context,
	rule *models.PaymentReceiveRule,
) (*models.PaymentReceiveRule, error) {
	err := r.client.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		var maxPosition *int
		if err := tx.Model(&models.PaymentReceiveRule{}).
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

func (r *PaymentReceiveRuleRepository) Update(
	ctx context.Context,
	rule *models.PaymentReceiveRule,
) (*models.PaymentReceiveRule, error) {
	if err := r.client.GetDB(ctx).Omit("Policy").Save(rule).Error; err != nil {
		return nil, err
	}
	return rule, nil
}

func (r *PaymentReceiveRuleRepository) DeleteById(ctx context.Context, id uuid.UUID) error {
	return r.client.GetDB(ctx).Delete(&models.PaymentReceiveRule{}, "id = ?", id).Error
}
