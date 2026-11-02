package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	// Not ON CONFLICT: MySQL's ON DUPLICATE KEY also fires on the token index and would overwrite another rule.
	db := r.client.GetDB(ctx)
	var count int64
	if err := db.Model(&models.PaymentReceiveRuleX402Config{}).
		Where("payment_receive_rule_id = ?", config.PaymentReceiveRuleId).
		Count(&count).
		Error; err != nil {
		return err
	}
	if count == 0 {
		return db.Omit("PaymentReceiveRule").Create(config).Error
	}
	return db.Model(&models.PaymentReceiveRuleX402Config{}).
		Where("payment_receive_rule_id = ?", config.PaymentReceiveRuleId).
		Select("policy_id", "network", "asset", "pay_to", "asset_transfer_method", "eip712_name", "eip712_version").
		Updates(config).
		Error
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

func (r *PaymentReceiveRuleX402ConfigRepository) GetByPolicyIdAndAsset(
	ctx context.Context,
	policyId uuid.UUID,
	network string,
	asset string,
) (*models.PaymentReceiveRuleX402Config, error) {
	var entity models.PaymentReceiveRuleX402Config
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
