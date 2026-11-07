package identity

import (
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const AgentCollectionsEndpoint = "/agent/collections"

// AgentCollectionsController is the owner-facing catalog of mintable
// collections, proxied from the runtime's on-chain manager read.
type AgentCollectionsController struct {
	listAgentCollections *use_cases.ListAgentCollections
	errorWriter          *mvc.ErrorWriter
}

func NewAgentCollectionsController(
	listAgentCollections *use_cases.ListAgentCollections,
	errorWriter *mvc.ErrorWriter,
) *AgentCollectionsController {
	return &AgentCollectionsController{listAgentCollections: listAgentCollections, errorWriter: errorWriter}
}

func (c *AgentCollectionsController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *AgentCollectionsController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(AgentCollectionsEndpoint, c.List)
}

// List returns the collections the on-chain manager governs.
// @Summary List agent collections
// @Tags AgentCollections
// @Produce json
// @Param request query requests.ListAgentCollectionsRequest true "Chain id, e.g. 80002"
// @Success 200 {object} responses.AgentCollectionsResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /agent/collections [get]
func (c *AgentCollectionsController) List(ctx fiber.Ctx) error {
	req := new(requests.ListAgentCollectionsRequest)
	if err := ctx.Bind().Query(req); err != nil {
		return c.errorWriter.Write(ctx, apperrors.NewBadRequestError("invalid_query", err.Error()))
	}
	resp, err := c.listAgentCollections.Execute(ctx.Context(), req)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}
