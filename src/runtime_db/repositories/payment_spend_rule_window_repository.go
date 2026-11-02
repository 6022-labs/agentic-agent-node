package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
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
) ([]models.PaymentSpendRuleWindow, error) {
	windows := []models.PaymentSpendRuleWindow{}
	err := r.client.GetDB(ctx).
		Where("payment_spend_rule_id = ?", ruleId).
		Order("window_seconds ASC").
		Find(&windows).
		Error
	if err != nil {
		return nil, err
	}
	return windows, nil
}

func (r *PaymentSpendRuleWindowRepository) ListByRuleIdForUpdate(
	ctx context.Context,
	ruleId uuid.UUID,
) ([]models.PaymentSpendRuleWindow, error) {
	query := r.client.GetDB(ctx).Where("payment_spend_rule_id = ?", ruleId).Order("window_seconds ASC")
	// sqlite has no row locks; DatabaseClient caps it at one connection instead.
	if r.client.GetDriver() != enums.DatabaseDriverSqlite {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	windows := []models.PaymentSpendRuleWindow{}
	if err := query.Find(&windows).Error; err != nil {
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
