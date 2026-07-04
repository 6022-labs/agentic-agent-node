package identity

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases"
	"github.com/gofiber/fiber/v2"
)

const AgentSoulEndpoint = "/agent/soul"

type AgentSoulController struct {
	getAgentSoul *use_cases.GetAgentSoul
}

func NewAgentSoulController(getAgentSoul *use_cases.GetAgentSoul) *AgentSoulController {
	return &AgentSoulController{getAgentSoul: getAgentSoul}
}

func (c *AgentSoulController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *AgentSoulController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(AgentSoulEndpoint, c.Get)
}

// Get returns the agent's soul — the immutable persona seeded at mint.
// @Summary Get agent soul
// @Tags Identity
// @Produce json
// @Success 200 {object} responses.AgentSoulResponse
// @Failure 404 {object} fiber.Map "identity_not_found"
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /agent/soul [get]
func (c *AgentSoulController) Get(ctx *fiber.Ctx) error {
	resp, err := c.getAgentSoul.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
