package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	runtime_common_repositories "github.com/6022-labs/agentic-agent-node/src/runtime_common/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	// Not ON CONFLICT: MySQL's ON DUPLICATE KEY also fires on the token index and would overwrite another rule.
	db := r.client.GetDB(ctx)
	var count int64
	if err := db.Model(&models.PaymentSpendRuleX402Config{}).
		Where("payment_spend_rule_id = ?", config.PaymentSpendRuleId).
		Count(&count).
		Error; err != nil {
		return err
	}
	if count == 0 {
		return db.Omit("PaymentSpendRule").Create(config).Error
	}
	return db.Model(&models.PaymentSpendRuleX402Config{}).
		Where("payment_spend_rule_id = ?", config.PaymentSpendRuleId).
		Select("policy_id", "network", "asset", "asset_transfer_method").
		Updates(config).
		Error
}

func (r *PaymentSpendRuleX402ConfigRepository) GetByRuleId(
	ctx context.Context,
	ruleId uuid.UUID,
	options ...runtime_common_repositories.ReadOption,
) (*models.PaymentSpendRuleX402Config, error) {
	var entity models.PaymentSpendRuleX402Config
	err := r.client.GetDB(ctx, options...).First(&entity, "payment_spend_rule_id = ?", ruleId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *PaymentSpendRuleX402ConfigRepository) GetByPolicyIdAndAsset(
	ctx context.Context,
	policyId uuid.UUID,
	network string,
	asset string,
) (*models.PaymentSpendRuleX402Config, error) {
	var entity models.PaymentSpendRuleX402Config
	err := r.client.GetDB(ctx).
		First(&entity, "policy_id = ? AND network = ? AND asset = ?", policyId, network, asset).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
