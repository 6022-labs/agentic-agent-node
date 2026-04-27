package gateway_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	McpConnectionsEndpoint    = "/config/mcp-connections"
	McpConnectionByIdEndpoint = "/config/mcp-connections/:id"
)

type McpConfigController struct {
	lister  *use_cases.ListMcpConnections
	creator *use_cases.CreateMcpConnection
	deleter *use_cases.DeleteMcpConnection
}

func NewMcpConfigController(
	lister *use_cases.ListMcpConnections,
	creator *use_cases.CreateMcpConnection,
	deleter *use_cases.DeleteMcpConnection,
) *McpConfigController {
	return &McpConfigController{lister: lister, creator: creator, deleter: deleter}
}

func (c *McpConfigController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *McpConfigController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(McpConnectionsEndpoint, c.List)
	app.Post(McpConnectionsEndpoint, c.Create)
	app.Delete(McpConnectionByIdEndpoint, c.Delete)
}

// List returns registered MCP connections.
// @Summary List MCP connections
// @Tags Config
// @Produce json
// @Success 200 {object} responses.ListMcpConnectionsResponse
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/mcp-connections [get]
func (c *McpConfigController) List(ctx *fiber.Ctx) error {
	res, err := c.lister.Execute(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"code": "runtime_unavailable", "message": err.Error()})
	}
	return ctx.JSON(res)
}

// Create registers a new MCP connection.
// @Summary Create MCP connection
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.CreateMcpConnectionRequest true "MCP connection"
// @Success 201 {object} responses.McpConnectionResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/mcp-connections [post]
func (c *McpConfigController) Create(ctx *fiber.Ctx) error {
	body := new(requests.CreateMcpConnectionRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	res, err := c.creator.Execute(ctx.Context(), body)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"code": "runtime_unavailable", "message": err.Error()})
	}
	return ctx.Status(fiber.StatusCreated).JSON(res)
}

// Delete removes an MCP connection by id.
// @Summary Delete MCP connection
// @Tags Config
// @Produce json
// @Param id path string true "MCP connection id"
// @Success 204 ""
// @Failure 400 {object} errors.ValidationError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/mcp-connections/{id} [delete]
func (c *McpConfigController) Delete(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.NewValidationError("id", "must be a valid uuid"))
	}
	if err := c.deleter.Execute(ctx.Context(), id); err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"code": "runtime_unavailable", "message": err.Error()})
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
