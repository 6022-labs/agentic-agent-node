package agent_reasoning_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v2"
)

const (
	InternalMCPConfigEndpoint      = "/internal/config/mcp"
	InternalMCPConfigToolsEndpoint = "/internal/config/mcp/tools"
	InternalMCPConfigByIdEndpoint  = "/internal/config/mcp/:id"
)

type MCPConfigController struct {
	listMCPConnections  *use_cases.ListMCPConnections
	getMCPConnection    *use_cases.GetMCPConnection
	createMCPConnection *use_cases.CreateMCPConnection
	updateMCPConnection *use_cases.UpdateMCPConnection
	deleteMCPConnection *use_cases.DeleteMCPConnection
	listMCPServerTools  *use_cases.DiscoverMCPServerTools
}

func NewMCPConfigController(
	listMCPConnections *use_cases.ListMCPConnections,
	getMCPConnection *use_cases.GetMCPConnection,
	createMCPConnection *use_cases.CreateMCPConnection,
	updateMCPConnection *use_cases.UpdateMCPConnection,
	deleteMCPConnection *use_cases.DeleteMCPConnection,
	listMCPServerTools *use_cases.DiscoverMCPServerTools,
) *MCPConfigController {
	return &MCPConfigController{
		listMCPConnections:  listMCPConnections,
		getMCPConnection:    getMCPConnection,
		createMCPConnection: createMCPConnection,
		updateMCPConnection: updateMCPConnection,
		deleteMCPConnection: deleteMCPConnection,
		listMCPServerTools:  listMCPServerTools,
	}
}

func (c *MCPConfigController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalMCPConfigEndpoint, c.List)
	app.Post(InternalMCPConfigEndpoint, c.Create)
	// Registered before the :id routes so "tools" is never parsed as a connection id.
	app.Post(InternalMCPConfigToolsEndpoint, c.DiscoverTools)
	app.Get(InternalMCPConfigByIdEndpoint, c.GetById)
	app.Put(InternalMCPConfigByIdEndpoint, c.Update)
	app.Delete(InternalMCPConfigByIdEndpoint, c.Delete)
}

// DiscoverTools asks an MCP server which tools it exposes.
// @Summary Internal discover MCP server tools
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.DiscoverMCPServerToolsRequest true "MCP server"
// @Success 200 {object} responses.MCPServerToolsResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 500 {object} fiber.Map "list_mcp_server_tools_failed"
// @Router /internal/config/mcp/tools [post]
func (c *MCPConfigController) DiscoverTools(ctx *fiber.Ctx) error {
	body := new(requests.DiscoverMCPServerToolsRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.listMCPServerTools.Execute(ctx.UserContext(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// List returns all of the agent's MCP connections.
// @Summary Internal list MCP connections
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.MCPConnectionListResponse
// @Failure 500 {object} fiber.Map "list_mcp_connections_failed"
// @Router /internal/config/mcp [get]
func (c *MCPConfigController) List(ctx *fiber.Ctx) error {
	resp, err := c.listMCPConnections.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Create adds a new MCP connection.
// @Summary Internal create MCP connection
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.CreateMCPConnectionRequest true "MCP connection"
// @Success 201 {object} responses.MCPConnectionResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} fiber.Map "create_mcp_connection_failed"
// @Router /internal/config/mcp [post]
func (c *MCPConfigController) Create(ctx *fiber.Ctx) error {
	body := new(requests.CreateMCPConnectionRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.createMCPConnection.Execute(ctx.UserContext(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

// GetById returns one MCP connection by id.
// @Summary Internal get MCP connection
// @Tags Internal
// @Produce json
// @Param id path string true "connection id"
// @Success 200 {object} responses.MCPConnectionResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 500 {object} fiber.Map "get_mcp_connection_failed"
// @Router /internal/config/mcp/{id} [get]
func (c *MCPConfigController) GetById(ctx *fiber.Ctx) error {
	request := new(requests.GetMCPConnectionRequest)
	if err := ctx.ParamsParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	resp, err := c.getMCPConnection.Execute(ctx.UserContext(), request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Update replaces an MCP connection by id.
// @Summary Internal update MCP connection
// @Tags Internal
// @Accept json
// @Produce json
// @Param id path string true "connection id"
// @Param request body requests.UpdateMCPConnectionRequest true "MCP connection"
// @Success 200 {object} responses.MCPConnectionResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 500 {object} fiber.Map "update_mcp_connection_failed"
// @Router /internal/config/mcp/{id} [put]
func (c *MCPConfigController) Update(ctx *fiber.Ctx) error {
	request := new(requests.UpdateMCPConnectionRequest)
	if err := ctx.ParamsParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.updateMCPConnection.Execute(ctx.UserContext(), request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete removes an MCP connection by id.
// @Summary Internal delete MCP connection
// @Tags Internal
// @Param id path string true "connection id"
// @Success 204
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} fiber.Map "delete_mcp_connection_failed"
// @Router /internal/config/mcp/{id} [delete]
func (c *MCPConfigController) Delete(ctx *fiber.Ctx) error {
	request := new(requests.DeleteMCPConnectionRequest)
	if err := ctx.ParamsParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := c.deleteMCPConnection.Execute(ctx.UserContext(), request); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
