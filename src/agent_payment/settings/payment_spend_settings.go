package settings

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const PaymentSpendSettingsKey = "payment.spend"

// PaymentSpendSettings bounds only the signing window; what may be spent lives in payment_spend_rules.
type PaymentSpendSettings struct {
	ValiditySeconds int64 `koanf:"validity_seconds"`
	ValidAfterSkew  int64 `koanf:"valid_after_skew_seconds"`
}

func NewPaymentSpendSettings(logger *zap.Logger, k *koanf.Koanf) *PaymentSpendSettings {
	settings := PaymentSpendSettings{
		ValiditySeconds: 300,
		ValidAfterSkew:  60,
	}
	if err := k.Unmarshal(PaymentSpendSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal spend settings", zap.Error(err))
	}

	if settings.ValiditySeconds <= 0 {
		logger.Warn("invalid payment.spend.validity_seconds, defaulting to 300")
		settings.ValiditySeconds = 300
	}
	if settings.ValidAfterSkew < 0 {
		logger.Warn("invalid payment.spend.valid_after_skew_seconds, defaulting to 60")
		settings.ValidAfterSkew = 60
	}

	return &settings
}
