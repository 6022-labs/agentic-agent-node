package core

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	"github.com/gofiber/fiber/v3"
)

const (
	AgentCardEndpoint = "/.well-known/agent-card.json"
	Agent6022Endpoint = "/.well-known/6022"
)

// WellKnownController serves the gateway's /.well-known/* discovery documents (static, unauthenticated).
type WellKnownController struct {
	getWellKnownAgentCard *use_cases.GetWellKnownAgentCard
	getWellKnownAgent6022 *use_cases.GetWellKnownAgent6022
}

func NewWellKnownController(
	getWellKnownAgentCard *use_cases.GetWellKnownAgentCard,
	getWellKnownAgent6022 *use_cases.GetWellKnownAgent6022,
) *WellKnownController {
	return &WellKnownController{
		getWellKnownAgentCard: getWellKnownAgentCard,
		getWellKnownAgent6022: getWellKnownAgent6022,
	}
}

func (c *WellKnownController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Get(AgentCardEndpoint, c.GetWellKnownAgentCard)
	app.Get(Agent6022Endpoint, c.GetWellKnownAgent6022)
}

func (c *WellKnownController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// GetWellKnownAgentCard serves the A2A v1.0 discovery card.
// @Summary A2A agent card
// @Description Open A2A v1.0 agent card (ProtoJSON); advertises SendMessage. Errors until the node is registered.
// @Tags A2A
// @Produce json
// @Success 200 {object} responses.WellKnownAgentCard
// @Failure 404 {object} errors.NotFoundError
// @Failure 502 {object} errors.UnavailableError
// @Router /.well-known/agent-card.json [get]
func (c *WellKnownController) GetWellKnownAgentCard(ctx fiber.Ctx) error {
	card, err := c.getWellKnownAgentCard.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(card)
}

// GetWellKnownAgent6022 serves the 6022-protocol agent card.
// @Summary 6022 agent card
// @Description Open 6022 discovery document; documents POST /responses and its bounded `instructions` overlay.
// @Tags 6022
// @Produce json
// @Success 200 {object} responses.WellKnownAgent6022
// @Failure 404 {object} errors.NotFoundError
// @Failure 502 {object} errors.UnavailableError
// @Router /.well-known/6022 [get]
func (c *WellKnownController) GetWellKnownAgent6022(ctx fiber.Ctx) error {
	card, err := c.getWellKnownAgent6022.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(card)
}
