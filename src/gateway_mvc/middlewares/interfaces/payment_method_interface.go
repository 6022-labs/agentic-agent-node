package interfaces

import (
	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/gofiber/fiber/v2"
)

// PaymentMethodInterface is one payment method's gate behavior. The dispatcher
// routes a request to the method that recognizes its payment proof, or to the
// default method (which advertises a 402) when none is present.
type PaymentMethodInterface interface {
	// Method is the method's wire identifier (e.g. "x402").
	Method() string
	// HasProof reports whether the request carries this method's payment.
	HasProof(ctx *fiber.Ctx) bool
	// Handle runs the method's gate: it verifies/settles a present payment and
	// calls the inner handler, or writes the 402 offer when none is present.
	Handle(ctx *fiber.Ctx, rules []models.ExternalPaymentRule) error
}
