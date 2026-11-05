package use_cases

import (
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
)

type RevokeAuth struct {
	refreshTokenManager gateway_interfaces.RefreshTokenManagerInterface
	refreshSessionStore gateway_interfaces.RefreshSessionStoreInterface
}

func NewRevokeAuth(
	refreshTokenManager gateway_interfaces.RefreshTokenManagerInterface,
	refreshSessionStore gateway_interfaces.RefreshSessionStoreInterface,
) *RevokeAuth {
	return &RevokeAuth{
		refreshTokenManager: refreshTokenManager,
		refreshSessionStore: refreshSessionStore,
	}
}

func (u *RevokeAuth) Execute(req *requests.RefreshTokenRequest) error {
	if req.RefreshToken == "" {
		return nil
	}
	claims, err := u.refreshTokenManager.Parse(req.RefreshToken)
	if err != nil {
		return nil
	}
	u.refreshSessionStore.Revoke(claims.SessionId, claims.SessionExpiresAt)
	return nil
}
