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

type PaymentRuleX402ConfigRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentRuleX402ConfigRepository(client *runtime_db.DatabaseClient) *PaymentRuleX402ConfigRepository {
	return &PaymentRuleX402ConfigRepository{client: client}
}

func (r *PaymentRuleX402ConfigRepository) Upsert(ctx context.Context, config *models.PaymentRuleX402Config) error {
	return r.client.GetDB().WithContext(ctx).
		Omit("PaymentRule").
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "payment_rule_id"}},
			UpdateAll: true,
		}).
		Create(config).Error
}

func (r *PaymentRuleX402ConfigRepository) GetByRuleId(
	ctx context.Context,
	ruleId uuid.UUID,
) (*models.PaymentRuleX402Config, error) {
	var entity models.PaymentRuleX402Config
	err := r.client.GetDB().WithContext(ctx).First(&entity, "payment_rule_id = ?", ruleId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
