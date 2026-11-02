package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
)

type PaymentReceiveRulePriceRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentReceiveRulePriceRepository(client *runtime_db.DatabaseClient) *PaymentReceiveRulePriceRepository {
	return &PaymentReceiveRulePriceRepository{client: client}
}

func (r *PaymentReceiveRulePriceRepository) ListByRuleId(
	ctx context.Context,
	ruleId uuid.UUID,
) ([]models.PaymentReceiveRulePrice, error) {
	prices := []models.PaymentReceiveRulePrice{}
	err := r.client.GetDB(ctx).
		Where("payment_receive_rule_id = ?", ruleId).
		Order("created_at ASC, id ASC").
		Find(&prices).
		Error
	if err != nil {
		return nil, err
	}
	return prices, nil
}

func (r *PaymentReceiveRulePriceRepository) Create(
	ctx context.Context,
	price *models.PaymentReceiveRulePrice,
) (*models.PaymentReceiveRulePrice, error) {
	if err := r.client.GetDB(ctx).Omit("PaymentReceiveRule").Create(price).Error; err != nil {
		return nil, err
	}
	return price, nil
}

func (r *PaymentReceiveRulePriceRepository) DeleteByRuleId(ctx context.Context, ruleId uuid.UUID) error {
	return r.client.GetDB(ctx).
		Where("payment_receive_rule_id = ?", ruleId).
		Delete(&models.PaymentReceiveRulePrice{}).
		Error
}
