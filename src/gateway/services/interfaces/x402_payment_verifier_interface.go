package interfaces

import (
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models/x402"
	"github.com/ethereum/go-ethereum/common"
)

// ErrX402AuthorizationExpired marks an EIP-3009 authorization outside its
// validity window. Contract sentinel — every impl must return this exact error.
var ErrX402AuthorizationExpired = errors.New("x402 authorization outside its validity window")

// X402PaymentVerifierInterface checks an x402 "exact" EVM payment locally,
// without any RPC call.
type X402PaymentVerifierInterface interface {
	// RecoverPayer checks recipient binding and validity window (not the
	// amount — the facilitator's concern) and returns the EIP-712 signer,
	// which must match the authorization's From.
	RecoverPayer(payload *x402.PaymentPayload, config *x402.X402Config) (common.Address, error)
}
