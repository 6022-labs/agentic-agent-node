package agent_reasoning_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const InternalAgentGuidanceEndpoint = "/internal/agent/guidance"

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

func (c *AgentGuidanceController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalAgentGuidanceEndpoint, c.Get)
	app.Put(InternalAgentGuidanceEndpoint, c.Set)
}

// Get returns the owner-editable guidance layer of the system prompt.
// @Summary Get agent guidance
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.AgentGuidanceResponse
// @Failure 500 {object} fiber.Map "guidance_lookup_failed"
// @Router /internal/agent/guidance [get]
func (c *AgentGuidanceController) Get(ctx fiber.Ctx) error {
	resp, err := c.getAgentGuidance.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Set replaces the agent's guidance.
// @Summary Set agent guidance
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.SetAgentGuidanceRequest true "Agent guidance"
// @Success 200 {object} responses.AgentGuidanceResponse
// @Failure 500 {object} fiber.Map "guidance_update_failed"
// @Router /internal/agent/guidance [put]
func (c *AgentGuidanceController) Set(ctx fiber.Ctx) error {
	body := new(requests.SetAgentGuidanceRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.setAgentGuidance.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
