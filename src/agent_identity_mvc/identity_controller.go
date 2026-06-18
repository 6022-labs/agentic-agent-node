package agent_identity_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v2"
)

const (
	InternalIdentityEndpoint           = "/internal/identity"
	InternalIdentityEnsPublishEndpoint = "/internal/identity/ens/publish"
)

type IdentityController struct {
	getAgentIdentity         *use_cases.GetAgentIdentity
	publishEnsRuntimeRecords *use_cases.PublishEnsRuntimeRecords
}

func NewIdentityController(
	getAgentIdentity *use_cases.GetAgentIdentity,
	publishEnsRuntimeRecords *use_cases.PublishEnsRuntimeRecords,
) *IdentityController {
	return &IdentityController{
		getAgentIdentity:         getAgentIdentity,
		publishEnsRuntimeRecords: publishEnsRuntimeRecords,
	}
}

func (c *IdentityController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalIdentityEndpoint, c.Get)
	app.Post(InternalIdentityEnsPublishEndpoint, c.PublishEns)
}

// Get resolves the agent's NFT identity.
// @Summary Internal identity lookup
// @Description Returns the agent's collection address + tokenId + ENS name. 404 when no NFT has been minted yet.
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.GetAgentIdentityResponse
// @Failure 404 {object} fiber.Map "identity_not_found"
// @Failure 500 {object} fiber.Map "identity_lookup_failed"
// @Router /internal/identity [get]
func (c *IdentityController) Get(ctx *fiber.Ctx) error {
	resp, err := c.getAgentIdentity.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// PublishEns re-publishes the agent's runtime ENS `url` record. Use when the
// URL changed or the boot-time publish failed (e.g. NFT not yet minted).
// @Summary Re-publish ENS runtime records
// @Description Writes the `url` text record for the agent's NFT via AgentEnsRegistry. Requires the NFT to already exist.
// @Tags Internal
// @Produce json
// @Success 204
// @Failure 409 {object} fiber.Map "identity_not_found"
// @Failure 500 {object} fiber.Map "ens_publish_failed"
// @Router /internal/identity/ens/publish [post]
func (c *IdentityController) PublishEns(ctx *fiber.Ctx) error {
	if err := c.publishEnsRuntimeRecords.Execute(ctx.UserContext()); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
