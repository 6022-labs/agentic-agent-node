package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

// GetPaymentFacilitators serves the configured network→facilitator mapping to the get_payment_facilitators self tool.
type GetPaymentFacilitators struct {
	paymentReceiveSettings *runtime_common_settings.PaymentReceiveSettings
}

func NewGetPaymentFacilitators(
	paymentReceiveSettings *runtime_common_settings.PaymentReceiveSettings,
) *GetPaymentFacilitators {
	return &GetPaymentFacilitators{paymentReceiveSettings: paymentReceiveSettings}
}

func (u *GetPaymentFacilitators) Execute(_ context.Context) (*responses.PaymentFacilitatorsResponse, error) {
	return responses.ToPaymentFacilitatorsResponse(u.paymentReceiveSettings.Entries()), nil
}
