package agent_reasoning_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
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

func (c *LLMConfigController) Get(ctx *fiber.Ctx) error {
	resp, err := c.getLLMConnection.Execute(ctx.Context())
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(resp)
}

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
		return writeError(ctx, err)
	}
	return ctx.JSON(resp)
}

func (c *LLMConfigController) Delete(ctx *fiber.Ctx) error {
	if err := c.deleteLLMConnection.Execute(ctx.Context()); err != nil {
		return writeError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
