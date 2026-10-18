package middlewares

import (
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/gofiber/fiber/v3"
)

// CallerContextMiddleware moves the gateway's caller id header onto the ctx; use cases validate it.
type CallerContextMiddleware struct{}

func NewCallerContextMiddleware() *CallerContextMiddleware {
	return &CallerContextMiddleware{}
}

func (m *CallerContextMiddleware) Handle(ctx fiber.Ctx) error {
	if callerId := ctx.Get(helpers.CallerIdHeader); callerId != "" {
		ctx.SetContext(helpers.ContextWithCallerId(ctx.Context(), callerId))
	}
	return ctx.Next()
}
