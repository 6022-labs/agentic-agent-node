package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/stretchr/testify/assert"
)

type WhenGettingPaymentFacilitatorsTestingSuite struct {
	sut *use_cases.GetPaymentFacilitators
}

func WhenGettingPaymentFacilitatorsBeforeEach(_ *testing.T) *WhenGettingPaymentFacilitatorsTestingSuite {
	paymentReceiveSettings := &runtime_common_settings.PaymentReceiveSettings{
		"eip155:80002": {BaseUrl: "https://x402-amoy.polygon.technology", MaxTimeoutSeconds: 90},
		"eip155:84532": {BaseUrl: "https://x402.org/facilitator", MaxTimeoutSeconds: 60},
	}

	return &WhenGettingPaymentFacilitatorsTestingSuite{
		sut: use_cases.NewGetPaymentFacilitators(paymentReceiveSettings),
	}
}

func TestWhenGettingPaymentFacilitators(t *testing.T) {
	t.Parallel()

	t.Run("Given the node has configured facilitators", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return them in declaration order", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentFacilitatorsBeforeEach(t)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, resp.Facilitators, 2) {
				assert.Equal(t, "eip155:80002", resp.Facilitators[0].Network)
				assert.Equal(t, "https://x402-amoy.polygon.technology", resp.Facilitators[0].BaseUrl)
				assert.Equal(t, 90, resp.Facilitators[0].MaxTimeoutSeconds)
				assert.Equal(t, "eip155:84532", resp.Facilitators[1].Network)
				assert.Equal(t, 60, resp.Facilitators[1].MaxTimeoutSeconds)
			}
		})
	})
}
