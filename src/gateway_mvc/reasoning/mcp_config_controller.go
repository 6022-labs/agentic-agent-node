package reasoning

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/gofiber/fiber/v2"
)

const (
	McpConnectionsEndpoint     = "/config/mcp-connections"
	McpConnectionToolsEndpoint = "/config/mcp-connections/tools"
	McpConnectionByIdEndpoint  = "/config/mcp-connections/:id"
)

type McpConfigController struct {
	listMcpConnections  *use_cases.ListMcpConnections
	createMcpConnection *use_cases.CreateMcpConnection
	updateMcpConnection *use_cases.UpdateMcpConnection
	deleteMcpConnection *use_cases.DeleteMcpConnection
	discoverMcpServerTools  *use_cases.DiscoverMcpServerTools
}

func NewMcpConfigController(
	listMcpConnections *use_cases.ListMcpConnections,
	createMcpConnection *use_cases.CreateMcpConnection,
	updateMcpConnection *use_cases.UpdateMcpConnection,
	deleteMcpConnection *use_cases.DeleteMcpConnection,
	discoverMcpServerTools *use_cases.DiscoverMcpServerTools,
) *McpConfigController {
	return &McpConfigController{
		listMcpConnections:  listMcpConnections,
		createMcpConnection: createMcpConnection,
		updateMcpConnection: updateMcpConnection,
		deleteMcpConnection: deleteMcpConnection,
		discoverMcpServerTools:  discoverMcpServerTools,
	}
}

func (c *McpConfigController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *McpConfigController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(McpConnectionsEndpoint, c.List)
	app.Post(McpConnectionsEndpoint, c.Create)
	// Registered before the :id routes so "tools" is never parsed as a connection id.
	app.Post(McpConnectionToolsEndpoint, c.DiscoverTools)
	app.Put(McpConnectionByIdEndpoint, c.Update)
	app.Delete(McpConnectionByIdEndpoint, c.Delete)
}

// DiscoverTools asks an MCP server which tools it exposes.
// @Summary Discover MCP server tools
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.DiscoverMcpServerToolsRequest true "MCP server"
// @Success 200 {object} responses.McpServerToolsResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/mcp-connections/tools [post]
func (c *McpConfigController) DiscoverTools(ctx *fiber.Ctx) error {
	body := new(requests.DiscoverMcpServerToolsRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.discoverMcpServerTools.Execute(ctx.UserContext(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Update edits an MCP connection by id.
// @Summary Update MCP connection
// @Tags Config
// @Accept json
// @Produce json
// @Param id path string true "MCP connection id"
// @Param request body requests.UpdateMcpConnectionRequest true "MCP connection"
// @Success 200 {object} responses.McpConnectionResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/mcp-connections/{id} [put]
func (c *McpConfigController) Update(ctx *fiber.Ctx) error {
	request := new(requests.UpdateMcpConnectionRequest)
	if err := ctx.ParamsParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.updateMcpConnection.Execute(ctx.UserContext(), request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
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
	resp, err := c.listMcpConnections.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
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
	resp, err := c.createMcpConnection.Execute(ctx.UserContext(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(resp)
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
	request := new(requests.DeleteMcpConnectionRequest)
	if err := ctx.ParamsParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := c.deleteMcpConnection.Execute(ctx.UserContext(), request); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
