package use_cases

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/responses"
)

// refreshReuseGrace lets concurrent tabs spend the same token without tripping reuse detection.
const refreshReuseGrace = 30 * time.Second

var (
	ErrRefreshTokenMissing = errors.New("refresh token missing")
	ErrRefreshTokenInvalid = errors.New("refresh token invalid or expired")
	ErrSessionRevoked      = errors.New("session revoked")
	ErrRefreshTokenReused  = errors.New("refresh token reused; session revoked")
)

type RefreshAuth struct {
	refreshTokenManager gateway_interfaces.RefreshTokenManagerInterface
	refreshSessionStore gateway_interfaces.RefreshSessionStoreInterface
	authorizer          common_interfaces.AuthorizerInterface
	jwtIssuer           gateway_interfaces.JwtIssuerInterface
	now                 func() time.Time
}

func NewRefreshAuth(
	refreshTokenManager gateway_interfaces.RefreshTokenManagerInterface,
	refreshSessionStore gateway_interfaces.RefreshSessionStoreInterface,
	authorizer common_interfaces.AuthorizerInterface,
	jwtIssuer gateway_interfaces.JwtIssuerInterface,
) *RefreshAuth {
	return &RefreshAuth{
		refreshTokenManager: refreshTokenManager,
		refreshSessionStore: refreshSessionStore,
		authorizer:          authorizer,
		jwtIssuer:           jwtIssuer,
		now:                 time.Now,
	}
}

func (u *RefreshAuth) Execute(
	ctx context.Context,
	req *requests.RefreshTokenRequest,
) (*responses.RefreshAuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, apperrors.NewUnauthorizedError("refresh_token_missing", ErrRefreshTokenMissing.Error())
	}
	claims, err := u.refreshTokenManager.Parse(req.RefreshToken)
	if err != nil {
		return nil, apperrors.NewUnauthorizedError("refresh_token_invalid", ErrRefreshTokenInvalid.Error())
	}
	// Authorize before spending so a transient failure never burns the token.
	authorized, err := u.authorizer.IsAuthorized(ctx, claims.Address)
	if err != nil {
		if errors.Is(err, common_interfaces.ErrIdentityNotFound) {
			return nil, apperrors.NewUnavailableError("identity_not_found", err.Error())
		}
		return nil, apperrors.NewInternalError("refresh_failed", err.Error())
	}
	if !authorized {
		u.refreshSessionStore.Revoke(claims.SessionId, claims.SessionExpiresAt)
		return nil, apperrors.NewUnauthorizedError("not_authorized", ErrAddressNotAuthorized.Error())
	}
	// Checked after the slow authorizer call so a concurrent logout is not missed.
	if u.refreshSessionStore.IsRevoked(claims.SessionId) {
		return nil, apperrors.NewUnauthorizedError("session_revoked", ErrSessionRevoked.Error())
	}

	successorId, err := u.refreshTokenManager.NewTokenId()
	if err != nil {
		return nil, apperrors.NewInternalError("refresh_failed", err.Error())
	}
	now := u.now()
	earlier, spent := u.refreshSessionStore.Spend(
		claims.SessionId,
		claims.TokenId,
		gateway_interfaces.SpentRefreshToken{SuccessorId: successorId, SpentAt: now},
		claims.ExpiresAt,
	)
	if !spent {
		if now.Sub(earlier.SpentAt) > refreshReuseGrace {
			u.refreshSessionStore.Revoke(claims.SessionId, claims.SessionExpiresAt)
			return nil, apperrors.NewUnauthorizedError("refresh_token_reused", ErrRefreshTokenReused.Error())
		}
		successorId = earlier.SuccessorId
	}

	token, expiresAt, err := u.jwtIssuer.Issue(claims.Address)
	if err != nil {
		return nil, apperrors.NewInternalError("refresh_failed", err.Error())
	}
	refreshToken, err := u.refreshTokenManager.Reissue(claims, successorId)
	if err != nil {
		return nil, apperrors.NewInternalError("refresh_failed", err.Error())
	}
	return &responses.RefreshAuthResponse{
		Token:            token,
		ExpiresAt:        expiresAt,
		RefreshToken:     refreshToken.Token,
		RefreshExpiresAt: refreshToken.ExpiresAt,
	}, nil
}
