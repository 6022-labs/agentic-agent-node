package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	runtime_common_repositories "github.com/6022-labs/agentic-agent-node/src/runtime_common/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
)

type PaymentSpendRuleWindowRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentSpendRuleWindowRepository(client *runtime_db.DatabaseClient) *PaymentSpendRuleWindowRepository {
	return &PaymentSpendRuleWindowRepository{client: client}
}

func (r *PaymentSpendRuleWindowRepository) ListByRuleId(
	ctx context.Context,
	ruleId uuid.UUID,
	options ...runtime_common_repositories.ReadOption,
) ([]models.PaymentSpendRuleWindow, error) {
	windows := []models.PaymentSpendRuleWindow{}
	err := r.client.GetDB(ctx, options...).
		Where("payment_spend_rule_id = ?", ruleId).
		Order("window_seconds ASC").
		Find(&windows).
		Error
	if err != nil {
		return nil, err
	}
	return windows, nil
}

func (r *PaymentSpendRuleWindowRepository) Create(
	ctx context.Context,
	window *models.PaymentSpendRuleWindow,
) (*models.PaymentSpendRuleWindow, error) {
	if err := r.client.GetDB(ctx).Omit("PaymentSpendRule").Create(window).Error; err != nil {
		return nil, err
	}
	return window, nil
}

func (r *PaymentSpendRuleWindowRepository) Update(
	ctx context.Context,
	window *models.PaymentSpendRuleWindow,
) (*models.PaymentSpendRuleWindow, error) {
	if err := r.client.GetDB(ctx).Omit("PaymentSpendRule").Save(window).Error; err != nil {
		return nil, err
	}
	return window, nil
}

func (r *PaymentSpendRuleWindowRepository) DeleteById(ctx context.Context, id uuid.UUID) error {
	return r.client.GetDB(ctx).Delete(&models.PaymentSpendRuleWindow{}, "id = ?", id).Error
}
