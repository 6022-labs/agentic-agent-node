package middlewares

import (
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/models/enums"
	"github.com/gofiber/fiber/v2"
)

type ReadyGateMiddleware struct {
	stateReader common_interfaces.BootstrapStateReaderInterface
}

func NewReadyGateMiddleware(stateReader common_interfaces.BootstrapStateReaderInterface) *ReadyGateMiddleware {
	return &ReadyGateMiddleware{stateReader: stateReader}
}

func (m *ReadyGateMiddleware) Handle(ctx *fiber.Ctx) error {
	state := m.stateReader.Get()
	if state.Status != enums.BootstrapStatusReady {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"code":    "not_ready",
			"message": "agent is not ready yet, current status: " + string(state.Status),
			"status":  state.Status,
		})
	}
	return ctx.Next()
}
