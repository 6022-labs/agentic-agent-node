package agent_identity_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const InternalAgentSoulEndpoint = "/internal/agent/soul"

type AgentSoulController struct {
	getAgentSoul *use_cases.GetAgentSoul
}

func NewAgentSoulController(getAgentSoul *use_cases.GetAgentSoul) *AgentSoulController {
	return &AgentSoulController{getAgentSoul: getAgentSoul}
}

func (c *AgentSoulController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalAgentSoulEndpoint, c.Get)
}

// Get returns the agent's soul — the immutable persona seeded at mint.
// @Summary Get agent soul
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.AgentSoulResponse
// @Failure 404 {object} fiber.Map "identity_not_found"
// @Failure 500 {object} fiber.Map "soul_lookup_failed"
// @Router /internal/agent/soul [get]
func (c *AgentSoulController) Get(ctx fiber.Ctx) error {
	resp, err := c.getAgentSoul.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
