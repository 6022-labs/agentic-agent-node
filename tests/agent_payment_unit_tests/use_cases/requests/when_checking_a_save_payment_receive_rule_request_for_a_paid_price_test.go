package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenCheckingASavePaymentReceiveRuleRequestForAPaidPrice(t *testing.T) {
	t.Parallel()

	t.Run("Given every price is free", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report no paid price", func(t *testing.T) {
			t.Parallel()

			req := &requests.SavePaymentReceiveRuleRequest{
				Prices: []requests.SavePaymentReceiveRulePriceRequest{{Price: "0"}},
			}

			assert.False(t, req.HasPaidPrice())
		})
	})

	t.Run("Given a paid price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report a paid price", func(t *testing.T) {
			t.Parallel()

			req := &requests.SavePaymentReceiveRuleRequest{
				Prices: []requests.SavePaymentReceiveRulePriceRequest{{Price: "0"}, {Price: "10000"}},
			}

			assert.True(t, req.HasPaidPrice())
		})
	})
}
