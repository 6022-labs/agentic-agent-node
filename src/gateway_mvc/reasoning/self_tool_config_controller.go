package reasoning

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const (
	SelfToolsEndpoint    = "/config/self-tools"
	SelfToolByIdEndpoint = "/config/self-tools/:id"
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

func (c *SelfToolConfigController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *SelfToolConfigController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(SelfToolsEndpoint, c.List)
	app.Put(SelfToolByIdEndpoint, c.Update)
}

// List returns every built-in self tool with its enabled flag.
// @Summary List self tools
// @Tags Config
// @Produce json
// @Success 200 {object} responses.ListSelfToolsResponse
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/self-tools [get]
func (c *SelfToolConfigController) List(ctx fiber.Ctx) error {
	resp, err := c.listSelfTools.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Update toggles one self tool by id.
// @Summary Update self tool
// @Tags Config
// @Accept json
// @Produce json
// @Param id path string true "self tool id"
// @Param request body requests.UpdateSelfToolRequest true "self tool"
// @Success 200 {object} responses.SelfToolResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/self-tools/{id} [put]
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
