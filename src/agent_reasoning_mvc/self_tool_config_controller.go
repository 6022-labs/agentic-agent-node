package agent_reasoning_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const (
	InternalSelfToolConfigEndpoint     = "/internal/config/self-tools"
	InternalSelfToolConfigByIdEndpoint = "/internal/config/self-tools/:id"
)

type SelfToolConfigController struct {
	listSelfTools  *use_cases.ListSelfTools
	updateSelfTool *use_cases.UpdateSelfTool
}

func NewSelfToolConfigController(
	listSelfTools *use_cases.ListSelfTools,
	updateSelfTool *use_cases.UpdateSelfTool,
) *SelfToolConfigController {
	return &SelfToolConfigController{
		listSelfTools:  listSelfTools,
		updateSelfTool: updateSelfTool,
	}
}

func (c *SelfToolConfigController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalSelfToolConfigEndpoint, c.List)
	app.Put(InternalSelfToolConfigByIdEndpoint, c.Update)
}

// List returns every self tool with its enabled flag.
// @Summary Internal list self tools
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.SelfToolListResponse
// @Failure 500 {object} errors.InternalError "list_failed or agent_lookup_failed"
// @Router /internal/config/self-tools [get]
func (c *SelfToolConfigController) List(ctx fiber.Ctx) error {
	resp, err := c.listSelfTools.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Update toggles one self tool by id.
// @Summary Internal update self tool
// @Tags Internal
// @Accept json
// @Produce json
// @Param id path string true "self tool id"
// @Param request body requests.UpdateSelfToolRequest true "self tool"
// @Success 200 {object} responses.SelfToolResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError "self_tool_not_found"
// @Failure 500 {object} errors.InternalError "get_failed, update_failed or agent_lookup_failed"
// @Router /internal/config/self-tools/{id} [put]
func (c *SelfToolConfigController) Update(ctx fiber.Ctx) error {
	request := new(requests.UpdateSelfToolRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := ctx.Bind().Body(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.updateSelfTool.Execute(ctx.Context(), request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
