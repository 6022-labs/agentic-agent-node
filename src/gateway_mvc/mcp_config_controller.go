package gateway_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/gofiber/fiber/v2"
)

const (
	McpConnectionsEndpoint   = "/config/mcp-connections"
	McpConnectionByIdEndpoint = "/config/mcp-connections/:id"
)

type McpConfigController struct{}

func NewMcpConfigController() *McpConfigController {
	return &McpConfigController{}
}

func (c *McpConfigController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *McpConfigController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(McpConnectionsEndpoint, c.List)
	app.Post(McpConnectionsEndpoint, c.Create)
	app.Delete(McpConnectionByIdEndpoint, c.Delete)
}

// List returns registered MCP connections (501).
// @Summary List MCP connections
// @Tags Config
// @Produce json
// @Success 501 {object} fiber.Map
// @Security BearerAuth
// @Router /config/mcp-connections [get]
func (c *McpConfigController) List(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("config.mcp.list"))
}

// Create registers a new MCP connection (501).
// @Summary Create MCP connection
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.CreateMcpConnectionRequest true "MCP connection"
// @Success 501 {object} fiber.Map
// @Failure 400 {object} errors.ValidationError
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
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("config.mcp.create"))
}

// Delete removes an MCP connection by id (501).
// @Summary Delete MCP connection
// @Tags Config
// @Produce json
// @Param id path string true "MCP connection id"
// @Success 501 {object} fiber.Map
// @Security BearerAuth
// @Router /config/mcp-connections/{id} [delete]
func (c *McpConfigController) Delete(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("config.mcp.delete"))
}
