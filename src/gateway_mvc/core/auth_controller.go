package core

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	"github.com/gofiber/fiber/v2"
)

const (
	AuthChallengeEndpoint = "/auth/challenge"
	AuthVerifyEndpoint    = "/auth/verify"
)

type AuthController struct {
	getAuthChallenge *use_cases.GetAuthChallenge
	verifyAuth       *use_cases.VerifyAuth
}

func NewAuthController(
	getAuthChallenge *use_cases.GetAuthChallenge,
	verifyAuth *use_cases.VerifyAuth,
) *AuthController {
	return &AuthController{
		getAuthChallenge: getAuthChallenge,
		verifyAuth:       verifyAuth,
	}
}

func (c *AuthController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Get(AuthChallengeEndpoint, c.Challenge)
	app.Post(AuthVerifyEndpoint, c.Verify)
}

func (c *AuthController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// Challenge issues a single-use nonce the caller must sign with their wallet.
// @Summary Auth challenge
// @Description Returns a fresh nonce + expiry. The caller signs the nonce with their wallet (EIP-191 personal_sign) and posts the result to /auth/verify.
// @Tags Auth
// @Produce json
// @Success 200 {object} responses.GetAuthChallengeResponse
// @Failure 500 {object} fiber.Map
// @Router /auth/challenge [get]
func (c *AuthController) Challenge(ctx *fiber.Ctx) error {
	resp, err := c.getAuthChallenge.Execute()
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Verify exchanges a signed nonce for a JWT.
// @Summary Auth verify
// @Description Verifies the EIP-191 signature, checks the recovered address against the NFT owner via the worker, and issues a HS256 JWT.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requests.VerifyAuthRequest true "Address + signature + nonce"
// @Success 200 {object} responses.VerifyAuthResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} fiber.Map "nonce_not_found | signature_mismatch | not_authorized"
// @Failure 502 {object} fiber.Map "identity_not_found"
// @Router /auth/verify [post]
func (c *AuthController) Verify(ctx *fiber.Ctx) error {
	body := new(requests.VerifyAuthRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.verifyAuth.Execute(ctx.UserContext(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
