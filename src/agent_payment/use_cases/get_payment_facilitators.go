package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

// GetPaymentFacilitators serves the configured network→facilitator mapping to
// the gateway's payment gate, which joins it with the rules to settle.
type GetPaymentFacilitators struct {
	paymentFacilitatorsSettings *runtime_common_settings.PaymentFacilitatorsSettings
}

func NewGetPaymentFacilitators(
	paymentFacilitatorsSettings *runtime_common_settings.PaymentFacilitatorsSettings,
) *GetPaymentFacilitators {
	return &GetPaymentFacilitators{paymentFacilitatorsSettings: paymentFacilitatorsSettings}
}

func (u *GetPaymentFacilitators) Execute(_ context.Context) (*responses.PaymentFacilitatorsResponse, error) {
	return responses.ToPaymentFacilitatorsResponse(u.paymentFacilitatorsSettings.Entries()), nil
}
