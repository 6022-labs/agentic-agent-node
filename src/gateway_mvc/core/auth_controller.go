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
	AuthRefreshEndpoint   = "/auth/refresh"
	AuthLogoutEndpoint    = "/auth/logout"
)

type AuthController struct {
	getAuthChallenge *use_cases.GetAuthChallenge
	verifyAuth       *use_cases.VerifyAuth
	refreshAuth      *use_cases.RefreshAuth
	revokeAuth       *use_cases.RevokeAuth
	authMetrics      metrics_interfaces.AuthMetricsInterface
}

func NewAuthController(
	getAuthChallenge *use_cases.GetAuthChallenge,
	verifyAuth *use_cases.VerifyAuth,
	refreshAuth *use_cases.RefreshAuth,
	revokeAuth *use_cases.RevokeAuth,
	authMetrics metrics_interfaces.AuthMetricsInterface,
) *AuthController {
	return &AuthController{
		getAuthChallenge: getAuthChallenge,
		verifyAuth:       verifyAuth,
		refreshAuth:      refreshAuth,
		revokeAuth:       revokeAuth,
		authMetrics:      authMetrics,
	}
}

func (c *AuthController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Get(AuthChallengeEndpoint, c.Challenge)
	app.Post(AuthVerifyEndpoint, c.Verify)
	// The access token has usually expired by now.
	app.Post(AuthRefreshEndpoint, c.Refresh)
	app.Post(AuthLogoutEndpoint, c.Logout)
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
// @Description Rebuilds the EIP-4361 message, verifies the EIP-191 signature, checks NFT ownership, issues a HS256 JWT and a refresh token.
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

// Refresh spends a refresh token and returns a fresh access JWT plus the rotated refresh token.
// @Summary Auth refresh
// @Description Spends the refresh token (valid jwt.refresh_idle_ttl after its last use, capped at jwt.refresh_ttl from sign-in), re-checks NFT ownership, issues a HS256 JWT and the next refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requests.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} responses.RefreshAuthResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "refresh_token_missing, refresh_token_invalid, session_revoked, refresh_token_reused or not_authorized"
// @Failure 500 {object} errors.InternalError "refresh_failed"
// @Failure 502 {object} errors.UnavailableError "identity_not_found"
// @Router /auth/refresh [post]
func (c *AuthController) Refresh(ctx fiber.Ctx) error {
	request := new(requests.RefreshTokenRequest)
	if err := ctx.Bind().Body(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.refreshAuth.Execute(ctx.Context(), request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Logout revokes the refresh token's session.
// @Summary Auth logout
// @Description Revokes the session behind the refresh token; idempotent.
// @Tags Auth
// @Accept json
// @Param request body requests.RefreshTokenRequest true "Refresh token"
// @Success 204
// @Failure 400 {object} errors.ValidationError
// @Router /auth/logout [post]
func (c *AuthController) Logout(ctx fiber.Ctx) error {
	request := new(requests.RefreshTokenRequest)
	if err := ctx.Bind().Body(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := c.revokeAuth.Execute(request); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
