package agent_reasoning_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc_result"
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
	return mvc_result.Write(ctx, c.getter.Execute(ctx.Context()))
}

func (c *LLMConfigController) Setup(ctx *fiber.Ctx) error {
	body := new(requests.SetupLLMConnectionRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	return mvc_result.Write(ctx, c.setupper.Execute(ctx.Context(), body))
}

func (c *LLMConfigController) Delete(ctx *fiber.Ctx) error {
	return mvc_result.Write(ctx, c.deleter.Execute(ctx.Context()))
}
