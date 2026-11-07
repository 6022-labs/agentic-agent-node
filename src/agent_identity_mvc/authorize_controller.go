package agent_identity_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const InternalAuthorizeEndpoint = "/internal/authorize"

type AuthorizeController struct {
	authorize   *use_cases.Authorize
	errorWriter *mvc.ErrorWriter
}

func NewAuthorizeController(authorize *use_cases.Authorize, errorWriter *mvc.ErrorWriter) *AuthorizeController {
	return &AuthorizeController{authorize: authorize, errorWriter: errorWriter}
}

func (c *AuthorizeController) RegisterRoutes(app fiber.Router) {
	app.Post(InternalAuthorizeEndpoint, c.Authorize)
}

// Authorize checks whether an address owns the agent's NFT.
// @Summary Internal authorize
// @Description Resolves the identity, fetches OwnerOf(tokenId), compares to the supplied address. Used by /auth/verify.
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.AuthorizeRequest true "Address to check"
// @Success 200 {object} responses.AuthorizeResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} errors.InternalError "authorize_failed"
// @Router /internal/authorize [post]
func (c *AuthorizeController) Authorize(ctx fiber.Ctx) error {
	body := new(requests.AuthorizeRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return c.errorWriter.Write(ctx, apperrors.NewBadRequestError("invalid_body", err.Error()))
	}
	resp, err := c.authorize.Execute(ctx.Context(), body)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}
