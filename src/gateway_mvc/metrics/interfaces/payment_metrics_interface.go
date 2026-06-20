package interfaces

import "context"

// PaymentMetricsInterface records x402 settlement outcomes at the gateway access gate.
type PaymentMetricsInterface interface {
	RecordSettled(ctx context.Context, network string)
	RecordRejected(ctx context.Context, network string)
}
