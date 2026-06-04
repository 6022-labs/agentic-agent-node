package agent_reasoning_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	InternalMCPConfigEndpoint     = "/internal/config/mcp"
	InternalMCPConfigByIdEndpoint = "/internal/config/mcp/:id"
)

type MCPConfigController struct {
	listMCPConnections  *use_cases.ListMCPConnections
	getMCPConnection    *use_cases.GetMCPConnection
	createMCPConnection *use_cases.CreateMCPConnection
	updateMCPConnection *use_cases.UpdateMCPConnection
	deleteMCPConnection *use_cases.DeleteMCPConnection
}

func NewMCPConfigController(
	listMCPConnections *use_cases.ListMCPConnections,
	getMCPConnection *use_cases.GetMCPConnection,
	createMCPConnection *use_cases.CreateMCPConnection,
	updateMCPConnection *use_cases.UpdateMCPConnection,
	deleteMCPConnection *use_cases.DeleteMCPConnection,
) *MCPConfigController {
	return &MCPConfigController{
		listMCPConnections:  listMCPConnections,
		getMCPConnection:    getMCPConnection,
		createMCPConnection: createMCPConnection,
		updateMCPConnection: updateMCPConnection,
		deleteMCPConnection: deleteMCPConnection,
	}
}

func (c *MCPConfigController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalMCPConfigEndpoint, c.List)
	app.Post(InternalMCPConfigEndpoint, c.Create)
	app.Get(InternalMCPConfigByIdEndpoint, c.GetById)
	app.Put(InternalMCPConfigByIdEndpoint, c.Update)
	app.Delete(InternalMCPConfigByIdEndpoint, c.Delete)
}

func (c *MCPConfigController) List(ctx *fiber.Ctx) error {
	resp, err := c.listMCPConnections.Execute(ctx.Context())
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(resp)
}

func (c *MCPConfigController) Create(ctx *fiber.Ctx) error {
	body := new(requests.CreateMCPConnectionRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	resp, err := c.createMCPConnection.Execute(ctx.Context(), body)
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

func (c *MCPConfigController) GetById(ctx *fiber.Ctx) error {
	id, idErr := parseConnectionId(ctx)
	if idErr != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(idErr)
	}
	resp, err := c.getMCPConnection.Execute(ctx.Context(), id)
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(resp)
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
	resp, err := c.updateMCPConnection.Execute(ctx.Context(), id, body)
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(resp)
}

func (c *MCPConfigController) Delete(ctx *fiber.Ctx) error {
	id, idErr := parseConnectionId(ctx)
	if idErr != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(idErr)
	}
	if err := c.deleteMCPConnection.Execute(ctx.Context(), id); err != nil {
		return writeError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func parseConnectionId(ctx *fiber.Ctx) (uuid.UUID, *errors.ValidationError) {
	parsed, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return uuid.Nil, errors.NewValidationError("id", "must be a valid uuid")
	}
	return parsed, nil
}
