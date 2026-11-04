package interfaces

import (
	"context"

	"go.uber.org/zap"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
)

// X402PaymentQuote is one accepted way to pay, as advertised in a 402.
type X402PaymentQuote struct {
	Network             string
	Asset               string
	PayTo               string
	Amount              string
	AssetTransferMethod enums.AssetTransferMethod
	Eip712Name          string
	Eip712Version       string
	// MaxTimeoutSeconds is the seller's longest signature life; 0 leaves the node's own validity.
	MaxTimeoutSeconds int64
}

// AuditFields is a payment's audit trail: what, to whom, how much; never the signature.
func (q *X402PaymentQuote) AuditFields() []zap.Field {
	return []zap.Field{
		zap.String("network", q.Network),
		zap.String("asset", q.Asset),
		zap.String("payTo", q.PayTo),
		zap.String("amount", q.Amount),
	}
}

// X402SignedPayment is a ready-to-send PAYMENT-SIGNATURE header and what it commits to.
type X402SignedPayment struct {
	PaymentSignature string
	Payer            string
	Network          string
	Asset            string
	PayTo            string
	Amount           string
	ValidBefore      int64
}

// X402PaymentSignerInterface signs an x402 "exact" payment with the node wallet.
type X402PaymentSignerInterface interface {
	Sign(ctx context.Context, quote *X402PaymentQuote) (*X402SignedPayment, error)
}
