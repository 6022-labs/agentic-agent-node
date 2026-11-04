package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingPaymentSpendRuleX402Config(t *testing.T) {
	t.Parallel()

	t.Run("Given a marshalled config", func(t *testing.T) {
		t.Parallel()

		t.Run("Should round-trip every wire field", func(t *testing.T) {
			t.Parallel()

			original := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithNetwork("eip155:80002").
				WithAssetTransferMethod(enums.AssetTransferMethodEip3009).
				Build()

			raw, err := original.Marshal()
			assert.NoError(t, err)

			var parsed models.PaymentSpendRuleX402Config
			assert.NoError(t, parsed.Parse(raw))

			assert.Equal(t, original.Network, parsed.Network)
			assert.Equal(t, original.Asset, parsed.Asset)
			assert.Equal(t, original.AssetTransferMethod, parsed.AssetTransferMethod)
		})

		t.Run("Should not carry the persistence fields onto the wire", func(t *testing.T) {
			t.Parallel()

			raw, err := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build().Marshal()
			assert.NoError(t, err)

			assert.NotContains(t, string(raw), "spendRuleId")
			assert.NotContains(t, string(raw), "PaymentSpendRule")
			assert.NotContains(t, string(raw), "olicyId")
		})
	})

	t.Run("Given an empty blob", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it as a validation error", func(t *testing.T) {
			t.Parallel()

			var config models.PaymentSpendRuleX402Config

			err := config.Parse(nil)

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config", validationError.Field)
		})
	})

	t.Run("Given a malformed blob", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it as a validation error", func(t *testing.T) {
			t.Parallel()

			var config models.PaymentSpendRuleX402Config

			err := config.Parse([]byte("not json"))

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
		})
	})
}
