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

type PaymentReceiveRuleX402ConfigRepository struct {
	client *runtime_db.DatabaseClient
}

func NewPaymentReceiveRuleX402ConfigRepository(
	client *runtime_db.DatabaseClient,
) *PaymentReceiveRuleX402ConfigRepository {
	return &PaymentReceiveRuleX402ConfigRepository{client: client}
}

func (r *PaymentReceiveRuleX402ConfigRepository) Upsert(
	ctx context.Context,
	config *models.PaymentReceiveRuleX402Config,
) error {
	return r.client.GetDB(ctx).
		Omit("PaymentReceiveRule").
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "payment_receive_rule_id"}},
			UpdateAll: true,
		}).
		Create(config).Error
}

func (r *PaymentReceiveRuleX402ConfigRepository) GetByRuleId(
	ctx context.Context,
	ruleId uuid.UUID,
) (*models.PaymentReceiveRuleX402Config, error) {
	var entity models.PaymentReceiveRuleX402Config
	err := r.client.GetDB(ctx).First(&entity, "payment_receive_rule_id = ?", ruleId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
