package gateway_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/gofiber/fiber/v2"
)

const LlmConfigEndpoint = "/config/llm"

type LlmConfigController struct{}

func NewLlmConfigController() *LlmConfigController {
	return &LlmConfigController{}
}

func (c *LlmConfigController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *LlmConfigController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(LlmConfigEndpoint, c.Get)
	app.Post(LlmConfigEndpoint, c.Set)
	app.Delete(LlmConfigEndpoint, c.Delete)
}

// Get returns the LLM provider configuration (501).
// @Summary Get LLM config
// @Tags Config
// @Produce json
// @Success 501 {object} fiber.Map
// @Security BearerAuth
// @Router /config/llm [get]
func (c *LlmConfigController) Get(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("config.llm.get"))
}

// Set upserts the LLM provider configuration (501).
// @Summary Set LLM config
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.SetLlmConfigRequest true "LLM config"
// @Success 501 {object} fiber.Map
// @Failure 400 {object} errors.ValidationError
// @Security BearerAuth
// @Router /config/llm [post]
func (c *LlmConfigController) Set(ctx *fiber.Ctx) error {
	body := new(requests.SetLlmConfigRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("config.llm.set"))
}

// Delete clears the LLM provider configuration (501).
// @Summary Delete LLM config
// @Tags Config
// @Produce json
// @Success 501 {object} fiber.Map
// @Security BearerAuth
// @Router /config/llm [delete]
func (c *LlmConfigController) Delete(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("config.llm.delete"))
}
