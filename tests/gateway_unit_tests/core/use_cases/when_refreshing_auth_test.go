package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	common_mocks "github.com/6022-labs/agentic-agent-node/tests/common_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	gateway_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/core_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenRefreshingAuthTestingSuite struct {
	sut *use_cases.RefreshAuth

	mockRefreshTokenManager *gateway_mocks.MockRefreshTokenManagerInterface
	mockRefreshSessionStore *gateway_mocks.MockRefreshSessionStoreInterface
	mockAuthorizer          *common_mocks.MockAuthorizerInterface
	mockJwtIssuer           *gateway_mocks.MockJwtIssuerInterface
}

func WhenRefreshingAuthBeforeEach(t *testing.T) *WhenRefreshingAuthTestingSuite {
	mockController := gomock.NewController(t)

	mockRefreshTokenManager := gateway_mocks.NewMockRefreshTokenManagerInterface(mockController)
	mockRefreshSessionStore := gateway_mocks.NewMockRefreshSessionStoreInterface(mockController)
	mockAuthorizer := common_mocks.NewMockAuthorizerInterface(mockController)
	mockJwtIssuer := gateway_mocks.NewMockJwtIssuerInterface(mockController)

	sut := use_cases.NewRefreshAuth(mockRefreshTokenManager, mockRefreshSessionStore, mockAuthorizer, mockJwtIssuer)

	return &WhenRefreshingAuthTestingSuite{
		sut: sut,

		mockRefreshTokenManager: mockRefreshTokenManager,
		mockRefreshSessionStore: mockRefreshSessionStore,
		mockAuthorizer:          mockAuthorizer,
		mockJwtIssuer:           mockJwtIssuer,
	}
}

func TestWhenRefreshingAuth(t *testing.T) {
	t.Parallel()

	address := common_tests.RandomEthAddress()
	claims := &gateway_interfaces.RefreshTokenClaims{
		Address:          address,
		SessionId:        "session-1",
		TokenId:          "token-1",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		SessionExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	request := &requests.RefreshTokenRequest{RefreshToken: "the-refresh-token"}
	expectAuthorizedSession := func(suite *WhenRefreshingAuthTestingSuite) {
		suite.mockRefreshTokenManager.EXPECT().Parse("the-refresh-token").Return(claims, nil)
		suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(true, nil)
		suite.mockRefreshSessionStore.EXPECT().IsRevoked("session-1").Return(false)
	}
	successorSpent := gomock.Cond(func(x any) bool {
		return x.(gateway_interfaces.SpentRefreshToken).SuccessorId == "token-2"
	})

	t.Run("Given no refresh token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return Unauthorized with refresh_token_missing", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.RefreshTokenRequest{})

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "refresh_token_missing", unauthorizedErr.Code)
		})
	})

	t.Run("Given the refresh token does not parse", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			suite.mockRefreshTokenManager.EXPECT().
				Parse("the-refresh-token").
				Return(nil, errors.New("token is expired"))
		}

		t.Run("Should return Unauthorized with refresh_token_invalid", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "refresh_token_invalid", unauthorizedErr.Code)
		})
	})

	t.Run("Given the authorizer reports identity-not-found", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			suite.mockRefreshTokenManager.EXPECT().Parse("the-refresh-token").Return(claims, nil)
			suite.mockAuthorizer.EXPECT().
				IsAuthorized(gomock.Any(), address).
				Return(false, common_interfaces.ErrIdentityNotFound)
		}

		t.Run("Should return Unavailable with identity_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "identity_not_found", unavailableErr.Code)
		})
	})

	t.Run("Given the authorizer fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			suite.mockRefreshTokenManager.EXPECT().Parse("the-refresh-token").Return(claims, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(false, errors.New("rpc down"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "refresh_failed", internalErr.Code)
			assert.Contains(t, internalErr.Message, "rpc down")
		})
	})

	t.Run("Given the authorizer rejects the address", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			suite.mockRefreshTokenManager.EXPECT().Parse("the-refresh-token").Return(claims, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(false, nil)
		}

		t.Run("Should return Unauthorized with not_authorized", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)
			suite.mockRefreshSessionStore.EXPECT().Revoke("session-1", claims.SessionExpiresAt)

			_, err := suite.sut.Execute(context.Background(), request)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "not_authorized", unauthorizedErr.Code)
		})

		t.Run("Should revoke the session", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)
			suite.mockRefreshSessionStore.EXPECT().Revoke("session-1", claims.SessionExpiresAt).Times(1)

			_, _ = suite.sut.Execute(context.Background(), request)
		})
	})

	t.Run("Given the session is revoked", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			suite.mockRefreshTokenManager.EXPECT().Parse("the-refresh-token").Return(claims, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(true, nil)
			suite.mockRefreshSessionStore.EXPECT().IsRevoked("session-1").Return(true)
		}

		t.Run("Should return Unauthorized with session_revoked", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "session_revoked", unauthorizedErr.Code)
		})
	})

	t.Run("Given the token id generation fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			expectAuthorizedSession(suite)
			suite.mockRefreshTokenManager.EXPECT().NewTokenId().Return("", errors.New("entropy exhausted"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "refresh_failed", internalErr.Code)
			assert.Contains(t, internalErr.Message, "entropy exhausted")
		})
	})

	t.Run("Given the token was spent outside the reuse grace window", func(t *testing.T) {
		t.Parallel()

		earlier := &gateway_interfaces.SpentRefreshToken{SuccessorId: "token-0", SpentAt: time.Now().Add(-time.Minute)}
		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			expectAuthorizedSession(suite)
			suite.mockRefreshTokenManager.EXPECT().NewTokenId().Return("token-2", nil)
			suite.mockRefreshSessionStore.EXPECT().
				Spend("session-1", "token-1", successorSpent, claims.ExpiresAt).
				Return(earlier, false)
		}

		t.Run("Should return Unauthorized with refresh_token_reused", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)
			suite.mockRefreshSessionStore.EXPECT().Revoke("session-1", claims.SessionExpiresAt)

			_, err := suite.sut.Execute(context.Background(), request)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "refresh_token_reused", unauthorizedErr.Code)
		})

		t.Run("Should revoke the session", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)
			suite.mockRefreshSessionStore.EXPECT().Revoke("session-1", claims.SessionExpiresAt).Times(1)

			_, _ = suite.sut.Execute(context.Background(), request)
		})
	})

	t.Run("Given the token was spent within the reuse grace window", func(t *testing.T) {
		t.Parallel()

		earlier := &gateway_interfaces.SpentRefreshToken{SuccessorId: "token-0", SpentAt: time.Now()}
		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			expectAuthorizedSession(suite)
			suite.mockRefreshTokenManager.EXPECT().NewTokenId().Return("token-2", nil)
			suite.mockRefreshSessionStore.EXPECT().
				Spend("session-1", "token-1", successorSpent, claims.ExpiresAt).
				Return(earlier, false)
			suite.mockJwtIssuer.EXPECT().Issue(address).Return("the-token", time.Now().Add(30*time.Minute), nil)
		}

		t.Run("Should reissue the earlier successor", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)
			suite.mockRefreshTokenManager.EXPECT().
				Reissue(claims, "token-0").
				Return(&gateway_interfaces.RefreshToken{Token: "the-successor-token", ExpiresAt: claims.ExpiresAt}, nil)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, "the-successor-token", resp.RefreshToken)
		})
	})

	t.Run("Given the JWT issuer fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			expectAuthorizedSession(suite)
			suite.mockRefreshTokenManager.EXPECT().NewTokenId().Return("token-2", nil)
			suite.mockRefreshSessionStore.EXPECT().
				Spend("session-1", "token-1", successorSpent, claims.ExpiresAt).
				Return(nil, true)
			suite.mockJwtIssuer.EXPECT().Issue(address).Return("", time.Time{}, errors.New("sign failed"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "refresh_failed", internalErr.Code)
			assert.Contains(t, internalErr.Message, "sign failed")
		})
	})

	t.Run("Given the refresh token manager fails to reissue", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			expectAuthorizedSession(suite)
			suite.mockRefreshTokenManager.EXPECT().NewTokenId().Return("token-2", nil)
			suite.mockRefreshSessionStore.EXPECT().
				Spend("session-1", "token-1", successorSpent, claims.ExpiresAt).
				Return(nil, true)
			suite.mockJwtIssuer.EXPECT().Issue(address).Return("the-token", time.Now(), nil)
			suite.mockRefreshTokenManager.EXPECT().Reissue(claims, "token-2").Return(nil, errors.New("sign failed"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "refresh_failed", internalErr.Code)
			assert.Contains(t, internalErr.Message, "sign failed")
		})
	})

	t.Run("Given an unspent token of a live session", func(t *testing.T) {
		t.Parallel()

		expectedExp := time.Now().Add(30 * time.Minute)
		initSuite := func(suite *WhenRefreshingAuthTestingSuite) {
			expectAuthorizedSession(suite)
			suite.mockRefreshTokenManager.EXPECT().NewTokenId().Return("token-2", nil)
			suite.mockRefreshSessionStore.EXPECT().
				Spend("session-1", "token-1", successorSpent, claims.ExpiresAt).
				Return(nil, true)
			suite.mockJwtIssuer.EXPECT().Issue(address).Return("the-token", expectedExp, nil)
			suite.mockRefreshTokenManager.EXPECT().
				Reissue(claims, "token-2").
				Return(&gateway_interfaces.RefreshToken{Token: "the-rotated-token", ExpiresAt: claims.ExpiresAt}, nil)
		}

		t.Run("Should return Ok with a JWT response", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, "the-token", resp.Token)
			assert.Equal(t, expectedExp, resp.ExpiresAt)
		})

		t.Run("Should return the rotated refresh token for the cookie", func(t *testing.T) {
			t.Parallel()

			suite := WhenRefreshingAuthBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, "the-rotated-token", resp.RefreshToken)
			assert.Equal(t, claims.ExpiresAt, resp.RefreshExpiresAt)
		})
	})
}
