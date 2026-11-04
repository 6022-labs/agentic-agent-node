package responses

import payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"

// X402PaymentSignatureResponse is a ready-to-send PAYMENT-SIGNATURE header.
type X402PaymentSignatureResponse struct {
	PaymentSignature string `json:"paymentSignature"`
	Payer            string `json:"payer"`
	Network          string `json:"network"`
	Asset            string `json:"asset"`
	PayTo            string `json:"payTo"`
	Amount           string `json:"amount"`
	ValidBefore      int64  `json:"validBefore"`
}

func (r *X402PaymentSignatureResponse) FromModel(signed *payment_interfaces.X402SignedPayment) {
	r.PaymentSignature = signed.PaymentSignature
	r.Payer = signed.Payer
	r.Network = signed.Network
	r.Asset = signed.Asset
	r.PayTo = signed.PayTo
	r.Amount = signed.Amount
	r.ValidBefore = signed.ValidBefore
}
