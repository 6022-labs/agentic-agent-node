package core

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	"github.com/gofiber/fiber/v2"
)

const AgentCardEndpoint = "/.well-known/agent-card.json"

// WellKnownController serves the gateway's /.well-known/* discovery documents (static, unauthenticated).
type WellKnownController struct {
	getWellKnownAgentCard *use_cases.GetWellKnownAgentCard
}

func NewWellKnownController(getWellKnownAgentCard *use_cases.GetWellKnownAgentCard) *WellKnownController {
	return &WellKnownController{getWellKnownAgentCard: getWellKnownAgentCard}
}

func (c *WellKnownController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Get(AgentCardEndpoint, c.GetWellKnownAgentCard)
}

func (c *WellKnownController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// GetWellKnownAgentCard serves the A2A v1.0 discovery card.
// @Summary A2A agent card
// @Description A2A v1.0 discovery document (ProtoJSON), open/unauthenticated. supportedInterfaces advertises the SendMessage endpoint. Requires agent identity, so it errors until the node is registered.
// @Tags A2A
// @Produce json
// @Success 200 {object} responses.WellKnownAgentCard
// @Failure 404 {object} errors.NotFoundError
// @Failure 502 {object} errors.UnavailableError
// @Router /.well-known/agent-card.json [get]
func (c *WellKnownController) GetWellKnownAgentCard(ctx *fiber.Ctx) error {
	card, err := c.getWellKnownAgentCard.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(card)
}
