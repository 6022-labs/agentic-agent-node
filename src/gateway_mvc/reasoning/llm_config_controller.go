package reasoning

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/gofiber/fiber/v2"
)

const LlmConfigEndpoint = "/config/llm"

type LlmConfigController struct {
	getLlmConfig    *use_cases.GetLlmConfig
	setLlmConfig    *use_cases.SetLlmConfig
	deleteLlmConfig *use_cases.DeleteLlmConfig
}

func NewLlmConfigController(
	getLlmConfig *use_cases.GetLlmConfig,
	setLlmConfig *use_cases.SetLlmConfig,
	deleteLlmConfig *use_cases.DeleteLlmConfig,
) *LlmConfigController {
	return &LlmConfigController{
		getLlmConfig:    getLlmConfig,
		setLlmConfig:    setLlmConfig,
		deleteLlmConfig: deleteLlmConfig,
	}
}

func (c *LlmConfigController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *LlmConfigController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(LlmConfigEndpoint, c.Get)
	app.Post(LlmConfigEndpoint, c.Set)
	app.Delete(LlmConfigEndpoint, c.Delete)
}

// Get returns the LLM provider configuration.
// @Summary Get LLM config
// @Tags Config
// @Produce json
// @Success 200 {object} responses.GetLlmConfigResponse
// @Failure 404 {object} fiber.Map "not_configured"
// @Security BearerAuth
// @Router /config/llm [get]
func (c *LlmConfigController) Get(ctx *fiber.Ctx) error {
	resp, err := c.getLlmConfig.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Set upserts the LLM provider configuration.
// @Summary Set LLM config
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.SetLlmConfigRequest true "LLM config"
// @Success 200 {object} responses.GetLlmConfigResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/llm [post]
func (c *LlmConfigController) Set(ctx *fiber.Ctx) error {
	body := new(requests.SetLlmConfigRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.setLlmConfig.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete clears the LLM provider configuration.
// @Summary Delete LLM config
// @Tags Config
// @Produce json
// @Success 204 ""
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/llm [delete]
func (c *LlmConfigController) Delete(ctx *fiber.Ctx) error {
	if err := c.deleteLlmConfig.Execute(ctx.Context()); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
