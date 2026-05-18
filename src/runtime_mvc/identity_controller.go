package runtime_mvc

import (
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/gofiber/fiber/v2"
)

const (
	InternalIdentityEndpoint           = "/internal/identity"
	InternalIdentityEnsPublishEndpoint = "/internal/identity/ens/publish"
)

type IdentityController struct {
	resolver     common_interfaces.AgentIdentityResolverInterface
	ensPublisher identity_interfaces.EnsRuntimeRecordsPublisherInterface
}

func NewIdentityController(
	resolver common_interfaces.AgentIdentityResolverInterface,
	ensPublisher identity_interfaces.EnsRuntimeRecordsPublisherInterface,
) *IdentityController {
	return &IdentityController{resolver: resolver, ensPublisher: ensPublisher}
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
	identity, err := c.resolver.Resolve(ctx.UserContext())
	if err != nil {
		if errors.Is(err, services.ErrIdentityNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "identity_not_found",
				"message": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "identity_lookup_failed",
			"message": err.Error(),
		})
	}
	return ctx.JSON(&responses.GetAgentIdentityResponse{
		CollectionAddress: identity.CollectionAddress,
		TokenId:           identity.TokenId,
		Name:              identity.Name,
	})
}

// PublishEns re-publishes the agent's runtime ENS text records (URL, kind).
// Use when the runtime URL has changed since the boot-time publish, or when
// the boot-time publish failed (e.g. the NFT had not yet been minted).
// @Summary Re-publish ENS runtime records
// @Description Writes `url` + `agent.kind` text records for the agent's NFT via AgentEnsRegistry. Requires the NFT to already exist.
// @Tags Internal
// @Produce json
// @Success 204
// @Failure 409 {object} fiber.Map "identity_not_found"
// @Failure 500 {object} fiber.Map "ens_publish_failed"
// @Router /internal/identity/ens/publish [post]
func (c *IdentityController) PublishEns(ctx *fiber.Ctx) error {
	if err := c.ensPublisher.PublishRuntimeRecords(ctx.UserContext()); err != nil {
		if errors.Is(err, services.ErrIdentityNotFound) {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "identity_not_found",
				"message": "agent NFT must be minted before ENS records can be published",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "ens_publish_failed",
			"message": err.Error(),
		})
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
