package runtime_mvc

import (
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/gofiber/fiber/v2"
)

const InternalIdentityEndpoint = "/internal/identity"

type IdentityController struct {
	resolver common_interfaces.AgentIdentityResolverInterface
}

func NewIdentityController(resolver common_interfaces.AgentIdentityResolverInterface) *IdentityController {
	return &IdentityController{resolver: resolver}
}

func (c *IdentityController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalIdentityEndpoint, c.Get)
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
