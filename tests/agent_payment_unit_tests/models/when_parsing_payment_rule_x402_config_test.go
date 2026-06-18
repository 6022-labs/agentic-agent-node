package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingPaymentRuleX402Config(t *testing.T) {
	t.Parallel()

	t.Run("Given a valid config blob", func(t *testing.T) {
		t.Parallel()

		t.Run("Should decode it and round-trip through Marshal", func(t *testing.T) {
			t.Parallel()

			config := &models.PaymentRuleX402Config{}
			assert.NoError(t, config.Parse([]byte(`{"network":"eip155:80002","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[".*"]}`)))
			assert.Equal(t, "eip155:80002", config.Network)
			assert.Equal(t, []string{".*"}, config.AddressPatterns)

			raw, err := config.Marshal()
			assert.NoError(t, err)

			reparsed := &models.PaymentRuleX402Config{}
			assert.NoError(t, reparsed.Parse(raw))
			assert.Equal(t, config, reparsed)
		})
	})

	t.Run("Given the blob has no persistence fields", func(t *testing.T) {
		t.Parallel()

		t.Run("Should omit them from the wire encoding", func(t *testing.T) {
			t.Parallel()

			config := &models.PaymentRuleX402Config{Network: "eip155:80002", AddressPatterns: []string{".*"}}

			raw, err := config.Marshal()
			assert.NoError(t, err)
			assert.NotContains(t, string(raw), "paymentRuleId")
			assert.NotContains(t, string(raw), "PaymentRule")
		})
	})

	t.Run("Given empty input", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			config := &models.PaymentRuleX402Config{}
			assert.Error(t, config.Parse(nil))
		})
	})

	t.Run("Given malformed JSON", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			config := &models.PaymentRuleX402Config{}
			assert.Error(t, config.Parse([]byte("{not-json")))
		})
	})
}
