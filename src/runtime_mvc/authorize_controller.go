package runtime_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc_result"
	"github.com/gofiber/fiber/v2"
)

const InternalAuthorizeEndpoint = "/internal/authorize"

type AuthorizeController struct {
	authorize *use_cases.Authorize
}

func NewAuthorizeController(authorize *use_cases.Authorize) *AuthorizeController {
	return &AuthorizeController{authorize: authorize}
}

func (c *AuthorizeController) RegisterRoutes(app fiber.Router) {
	app.Post(InternalAuthorizeEndpoint, c.Authorize)
}

// Authorize checks whether an address owns the agent's NFT.
// @Summary Internal authorize
// @Description Resolves the agent identity, fetches OwnerOf(tokenId), compares to the supplied address. Used by gateway during /auth/verify.
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.AuthorizeRequest true "Address to check"
// @Success 200 {object} responses.AuthorizeResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} fiber.Map "authorize_failed"
// @Router /internal/authorize [post]
func (c *AuthorizeController) Authorize(ctx *fiber.Ctx) error {
	body := new(requests.AuthorizeRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	return mvc_result.Write(ctx, c.authorize.Execute(ctx.UserContext(), body))
}
