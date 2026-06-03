package agent_reasoning_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc_result"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	InternalMCPConfigEndpoint     = "/internal/config/mcp"
	InternalMCPConfigByIdEndpoint = "/internal/config/mcp/:id"
)

type MCPConfigController struct {
	lister  *use_cases.ListMCPConnections
	getter  *use_cases.GetMCPConnection
	creator *use_cases.CreateMCPConnection
	updater *use_cases.UpdateMCPConnection
	deleter *use_cases.DeleteMCPConnection
}

func NewMCPConfigController(
	lister *use_cases.ListMCPConnections,
	getter *use_cases.GetMCPConnection,
	creator *use_cases.CreateMCPConnection,
	updater *use_cases.UpdateMCPConnection,
	deleter *use_cases.DeleteMCPConnection,
) *MCPConfigController {
	return &MCPConfigController{lister: lister, getter: getter, creator: creator, updater: updater, deleter: deleter}
}

func (c *MCPConfigController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalMCPConfigEndpoint, c.List)
	app.Post(InternalMCPConfigEndpoint, c.Create)
	app.Get(InternalMCPConfigByIdEndpoint, c.GetById)
	app.Put(InternalMCPConfigByIdEndpoint, c.Update)
	app.Delete(InternalMCPConfigByIdEndpoint, c.Delete)
}

func (c *MCPConfigController) List(ctx *fiber.Ctx) error {
	return mvc_result.Write(ctx, c.lister.Execute())
}

func (c *MCPConfigController) Create(ctx *fiber.Ctx) error {
	body := new(requests.CreateMCPConnectionRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	return mvc_result.Write(ctx, c.creator.Execute(ctx.Context(), body))
}

func (c *MCPConfigController) GetById(ctx *fiber.Ctx) error {
	id, idErr := parseConnectionId(ctx)
	if idErr != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(idErr)
	}
	return mvc_result.Write(ctx, c.getter.Execute(ctx.Context(), id))
}

func (c *MCPConfigController) Update(ctx *fiber.Ctx) error {
	id, idErr := parseConnectionId(ctx)
	if idErr != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(idErr)
	}
	body := new(requests.UpdateMCPConnectionRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	return mvc_result.Write(ctx, c.updater.Execute(ctx.Context(), id, body))
}

func (c *MCPConfigController) Delete(ctx *fiber.Ctx) error {
	id, idErr := parseConnectionId(ctx)
	if idErr != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(idErr)
	}
	return mvc_result.Write(ctx, c.deleter.Execute(ctx.Context(), id))
}

func parseConnectionId(ctx *fiber.Ctx) (uuid.UUID, *errors.ValidationError) {
	parsed, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return uuid.Nil, errors.NewValidationError("id", "must be a valid uuid")
	}
	return parsed, nil
}
