package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentSpendRuleX402ConfigRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentSpendRuleX402ConfigRepository(
	client *runtime_db.DatabaseClient,
) *PaymentSpendRuleX402ConfigRepository {
	return &PaymentSpendRuleX402ConfigRepository{client: client}
}

func (r *PaymentSpendRuleX402ConfigRepository) Upsert(
	ctx context.Context,
	config *models.PaymentSpendRuleX402Config,
) error {
	return r.client.GetDB(ctx).
		Omit("PaymentSpendRule").
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "payment_spend_rule_id"}},
			UpdateAll: true,
		}).
		Create(config).Error
}

func (r *PaymentSpendRuleX402ConfigRepository) GetByRuleId(
	ctx context.Context,
	ruleId uuid.UUID,
) (*models.PaymentSpendRuleX402Config, error) {
	var entity models.PaymentSpendRuleX402Config
	err := r.client.GetDB(ctx).First(&entity, "payment_spend_rule_id = ?", ruleId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
