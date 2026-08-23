package interfaces

import (
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/gofiber/fiber/v3"
)

// AccessMethodInterface is one link in the gate's access chain.
type AccessMethodInterface interface {
	// TryHandle handles the request (returns true) or defers to the next link.
	TryHandle(ctx fiber.Ctx, rules []models.ExternalPaymentRule) (handled bool, err error)
	// Priority orders links in the chain, lowest first.
	Priority() int
}
