package agent_reasoning_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v2"
)

const InternalLLMConfigEndpoint = "/internal/config/llm"

type LLMConfigController struct {
	getLLMConnection    *use_cases.GetLLMConnection
	setupLLMConnection  *use_cases.SetupLLMConnection
	deleteLLMConnection *use_cases.DeleteLLMConnection
}

func NewLLMConfigController(
	getLLMConnection *use_cases.GetLLMConnection,
	setupLLMConnection *use_cases.SetupLLMConnection,
	deleteLLMConnection *use_cases.DeleteLLMConnection,
) *LLMConfigController {
	return &LLMConfigController{
		getLLMConnection:    getLLMConnection,
		setupLLMConnection:  setupLLMConnection,
		deleteLLMConnection: deleteLLMConnection,
	}
}

func (c *LLMConfigController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalLLMConfigEndpoint, c.Get)
	app.Post(InternalLLMConfigEndpoint, c.Setup)
	app.Delete(InternalLLMConfigEndpoint, c.Delete)
}

// Get returns the agent's single LLM connection.
// @Summary Internal get LLM connection
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.LLMConnectionResponse
// @Failure 404 {object} errors.NotFoundError
// @Failure 500 {object} fiber.Map "get_llm_connection_failed"
// @Router /internal/config/llm [get]
func (c *LLMConfigController) Get(ctx *fiber.Ctx) error {
	resp, err := c.getLLMConnection.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Setup creates or replaces the agent's LLM connection.
// @Summary Internal setup LLM connection
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.SetupLLMConnectionRequest true "LLM connection"
// @Success 200 {object} responses.LLMConnectionResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} fiber.Map "setup_llm_connection_failed"
// @Router /internal/config/llm [post]
func (c *LLMConfigController) Setup(ctx *fiber.Ctx) error {
	body := new(requests.SetupLLMConnectionRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	resp, err := c.setupLLMConnection.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete removes the agent's LLM connection.
// @Summary Internal delete LLM connection
// @Tags Internal
// @Success 204
// @Failure 500 {object} fiber.Map "delete_llm_connection_failed"
// @Router /internal/config/llm [delete]
func (c *LLMConfigController) Delete(ctx *fiber.Ctx) error {
	if err := c.deleteLLMConnection.Execute(ctx.Context()); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
