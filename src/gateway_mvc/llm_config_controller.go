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

func (c *LlmConfigController) Get(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("config.llm.get"))
}

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

func (c *LlmConfigController) Delete(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("config.llm.delete"))
}
