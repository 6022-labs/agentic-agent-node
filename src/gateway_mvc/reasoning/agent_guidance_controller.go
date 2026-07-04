package reasoning

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/gofiber/fiber/v2"
)

const AgentGuidanceEndpoint = "/agent/guidance"

type AgentGuidanceController struct {
	getAgentGuidance *use_cases.GetAgentGuidance
	setAgentGuidance *use_cases.SetAgentGuidance
}

func NewAgentGuidanceController(
	getAgentGuidance *use_cases.GetAgentGuidance,
	setAgentGuidance *use_cases.SetAgentGuidance,
) *AgentGuidanceController {
	return &AgentGuidanceController{
		getAgentGuidance: getAgentGuidance,
		setAgentGuidance: setAgentGuidance,
	}
}

func (c *AgentGuidanceController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *AgentGuidanceController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(AgentGuidanceEndpoint, c.Get)
	app.Put(AgentGuidanceEndpoint, c.Set)
}

// Get returns the owner-editable guidance layer of the system prompt.
// @Summary Get agent guidance
// @Tags Config
// @Produce json
// @Success 200 {object} responses.AgentGuidanceResponse
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /agent/guidance [get]
func (c *AgentGuidanceController) Get(ctx *fiber.Ctx) error {
	resp, err := c.getAgentGuidance.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Set replaces the agent's guidance.
// @Summary Set agent guidance
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.SetAgentGuidanceRequest true "Agent guidance"
// @Success 200 {object} responses.AgentGuidanceResponse
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /agent/guidance [put]
func (c *AgentGuidanceController) Set(ctx *fiber.Ctx) error {
	body := new(requests.SetAgentGuidanceRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.setAgentGuidance.Execute(ctx.UserContext(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
