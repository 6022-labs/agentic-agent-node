package interfaces

import (
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/ethereum/go-ethereum/common"
)

// ErrX402AuthorizationExpired is the sentinel every impl must return for an expired window.
var ErrX402AuthorizationExpired = errors.New("x402 authorization outside its validity window")

// X402PaymentVerifierInterface checks an x402 "exact" EVM payment locally, without any RPC call.
type X402PaymentVerifierInterface interface {
	// RecoverPayer checks recipient and validity window, not the amount; the signer must match From.
	RecoverPayer(payload *x402.PaymentPayload, config *models.PaymentReceiveRuleX402Config) (common.Address, error)
}
