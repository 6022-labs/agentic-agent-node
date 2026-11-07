package reasoning

import (
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const AgentGuidanceEndpoint = "/agent/guidance"

type AgentGuidanceController struct {
	getAgentGuidance *use_cases.GetAgentGuidance
	setAgentGuidance *use_cases.SetAgentGuidance
	errorWriter      *mvc.ErrorWriter
}

func NewAgentGuidanceController(
	getAgentGuidance *use_cases.GetAgentGuidance,
	setAgentGuidance *use_cases.SetAgentGuidance,
	errorWriter *mvc.ErrorWriter,
) *AgentGuidanceController {
	return &AgentGuidanceController{
		getAgentGuidance: getAgentGuidance,
		setAgentGuidance: setAgentGuidance,
		errorWriter:      errorWriter,
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
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /agent/guidance [get]
func (c *AgentGuidanceController) Get(ctx fiber.Ctx) error {
	resp, err := c.getAgentGuidance.Execute(ctx.Context())
	if err != nil {
		return c.errorWriter.Write(ctx, err)
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
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /agent/guidance [put]
func (c *AgentGuidanceController) Set(ctx fiber.Ctx) error {
	body := new(requests.SetAgentGuidanceRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return c.errorWriter.Write(ctx, apperrors.NewBadRequestError("invalid_body", err.Error()))
	}
	resp, err := c.setAgentGuidance.Execute(ctx.Context(), body)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}
