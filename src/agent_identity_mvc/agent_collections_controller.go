package agent_identity_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v2"
)

const InternalAgentCollectionsEndpoint = "/internal/agent/collections"

// AgentCollectionsController exposes the on-chain collections catalog; the
// gateway proxies the owner's authenticated calls here.
type AgentCollectionsController struct {
	listAgentCollections *use_cases.ListAgentCollections
}

func NewAgentCollectionsController(
	listAgentCollections *use_cases.ListAgentCollections,
) *AgentCollectionsController {
	return &AgentCollectionsController{listAgentCollections: listAgentCollections}
}

func (c *AgentCollectionsController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalAgentCollectionsEndpoint, c.List)
}

// List returns the mintable collections from the on-chain manager.
// @Summary List agent collections
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.AgentCollectionsResponse
// @Failure 502 {object} fiber.Map "collections_read_failed"
// @Router /internal/agent/collections [get]
func (c *AgentCollectionsController) List(ctx *fiber.Ctx) error {
	resp, err := c.listAgentCollections.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
