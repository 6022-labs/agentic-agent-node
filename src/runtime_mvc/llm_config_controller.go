package runtime_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/gofiber/fiber/v2"
)

const InternalLLMConfigEndpoint = "/internal/config/llm"

type LLMConfigController struct {
	getter   *use_cases.GetLLMConnection
	setupper *use_cases.SetupLLMConnection
	deleter  *use_cases.DeleteLLMConnection
}

func NewLLMConfigController(
	getter *use_cases.GetLLMConnection,
	setupper *use_cases.SetupLLMConnection,
	deleter *use_cases.DeleteLLMConnection,
) *LLMConfigController {
	return &LLMConfigController{getter: getter, setupper: setupper, deleter: deleter}
}

func (c *LLMConfigController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalLLMConfigEndpoint, c.Get)
	app.Post(InternalLLMConfigEndpoint, c.Setup)
	app.Delete(InternalLLMConfigEndpoint, c.Delete)
}

func (c *LLMConfigController) Get(ctx *fiber.Ctx) error {
	connection, err := c.getter.Execute()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "get_failed", "message": err.Error()})
	}
	if connection == nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}
	return ctx.JSON(responses.ToLLMConnectionResponse(connection))
}

func (c *LLMConfigController) Setup(ctx *fiber.Ctx) error {
	body := new(requests.SetupLLMConnectionRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	connection, err := c.setupper.Execute(ctx.Context(), body)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "setup_failed", "message": err.Error()})
	}
	return ctx.JSON(responses.ToLLMConnectionResponse(connection))
}

func (c *LLMConfigController) Delete(ctx *fiber.Ctx) error {
	if err := c.deleter.Execute(ctx.Context()); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "delete_failed", "message": err.Error()})
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
