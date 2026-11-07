package identity

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases"
	"github.com/gofiber/fiber/v3"
)

const AgentSoulEndpoint = "/agent/soul"

type AgentSoulController struct {
	getAgentSoul *use_cases.GetAgentSoul
	errorWriter  *mvc.ErrorWriter
}

func NewAgentSoulController(getAgentSoul *use_cases.GetAgentSoul, errorWriter *mvc.ErrorWriter) *AgentSoulController {
	return &AgentSoulController{getAgentSoul: getAgentSoul, errorWriter: errorWriter}
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
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 404 {object} errors.NotFoundError "identity_not_found"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /agent/soul [get]
func (c *AgentSoulController) Get(ctx fiber.Ctx) error {
	resp, err := c.getAgentSoul.Execute(ctx.Context())
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}
