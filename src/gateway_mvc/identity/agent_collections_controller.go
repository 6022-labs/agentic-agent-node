package identity

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases"
	"github.com/gofiber/fiber/v2"
)

const AgentCollectionsEndpoint = "/agent/collections"

// AgentCollectionsController is the owner-facing catalog of mintable
// collections, proxied from the runtime's on-chain manager read.
type AgentCollectionsController struct {
	listAgentCollections *use_cases.ListAgentCollections
}

func NewAgentCollectionsController(
	listAgentCollections *use_cases.ListAgentCollections,
) *AgentCollectionsController {
	return &AgentCollectionsController{listAgentCollections: listAgentCollections}
}

func (c *AgentCollectionsController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *AgentCollectionsController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(AgentCollectionsEndpoint, c.List)
}

// List returns the collections the on-chain manager governs.
// @Summary List agent collections
// @Tags AgentCollections
// @Produce json
// @Success 200 {object} responses.AgentCollectionsResponse
// @Security BearerAuth
// @Router /agent/collections [get]
func (c *AgentCollectionsController) List(ctx *fiber.Ctx) error {
	resp, err := c.listAgentCollections.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
