package core

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/metrics/interfaces"
	"github.com/gofiber/fiber/v3"
)

const (
	AuthChallengeEndpoint = "/auth/challenge"
	AuthVerifyEndpoint    = "/auth/verify"
)

type AuthController struct {
	getAuthChallenge *use_cases.GetAuthChallenge
	verifyAuth       *use_cases.VerifyAuth
	authMetrics      metrics_interfaces.AuthMetricsInterface
}

func NewAuthController(
	getAuthChallenge *use_cases.GetAuthChallenge,
	verifyAuth *use_cases.VerifyAuth,
	authMetrics metrics_interfaces.AuthMetricsInterface,
) *AuthController {
	return &AuthController{
		getAuthChallenge: getAuthChallenge,
		verifyAuth:       verifyAuth,
		authMetrics:      authMetrics,
	}
}

func (c *AuthController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Get(AuthChallengeEndpoint, c.Challenge)
	app.Post(AuthVerifyEndpoint, c.Verify)
}

func (c *AuthController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// Challenge issues a single-use EIP-4361 message the caller must sign with their wallet.
// @Summary Auth challenge
// @Description Returns a fresh SIWE (EIP-4361) message + expiry; sign it verbatim (EIP-191) and post to /auth/verify.
// @Tags Auth
// @Produce json
// @Param address query string true "wallet address the message is issued for"
// @Param chainId query int true "chain id of the connected wallet"
// @Success 200 {object} responses.GetAuthChallengeResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} errors.InternalError "challenge_failed"
// @Router /auth/challenge [get]
func (c *AuthController) Challenge(ctx fiber.Ctx) error {
	request := new(requests.GetAuthChallengeRequest)
	if err := ctx.Bind().Query(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_query", "message": err.Error()})
	}
	resp, err := c.getAuthChallenge.Execute(request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Verify exchanges a signed SIWE message for a JWT.
// @Summary Auth verify
// @Description Rebuilds the EIP-4361 message, verifies the EIP-191 signature, checks NFT ownership, issues a HS256 JWT.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requests.VerifyAuthRequest true "Address + signature"
// @Success 200 {object} responses.VerifyAuthResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "challenge_not_found, signature_mismatch or not_authorized"
// @Failure 500 {object} errors.InternalError "verify_failed"
// @Failure 502 {object} errors.UnavailableError "identity_not_found"
// @Router /auth/verify [post]
func (c *AuthController) Verify(ctx fiber.Ctx) error {
	body := new(requests.VerifyAuthRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.verifyAuth.Execute(ctx.Context(), body)
	if err != nil {
		c.authMetrics.RecordAuthFailure(ctx.Context())
		return mvc.WriteError(ctx, err)
	}
	c.authMetrics.RecordAuthSuccess(ctx.Context())
	return ctx.JSON(resp)
}
