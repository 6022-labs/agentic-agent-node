package use_cases_test

import (
	"errors"
	"testing"
	"time"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	gateway_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/core_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenRevokingAuthTestingSuite struct {
	sut *use_cases.RevokeAuth

	mockRefreshTokenManager *gateway_mocks.MockRefreshTokenManagerInterface
	mockRefreshSessionStore *gateway_mocks.MockRefreshSessionStoreInterface
}

func WhenRevokingAuthBeforeEach(t *testing.T) *WhenRevokingAuthTestingSuite {
	mockController := gomock.NewController(t)

	mockRefreshTokenManager := gateway_mocks.NewMockRefreshTokenManagerInterface(mockController)
	mockRefreshSessionStore := gateway_mocks.NewMockRefreshSessionStoreInterface(mockController)

	return &WhenRevokingAuthTestingSuite{
		sut: use_cases.NewRevokeAuth(mockRefreshTokenManager, mockRefreshSessionStore),

		mockRefreshTokenManager: mockRefreshTokenManager,
		mockRefreshSessionStore: mockRefreshSessionStore,
	}
}

func TestWhenRevokingAuth(t *testing.T) {
	t.Parallel()

	request := &requests.RefreshTokenRequest{RefreshToken: "the-refresh-token"}

	t.Run("Given no refresh token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should succeed without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenRevokingAuthBeforeEach(t)

			err := suite.sut.Execute(&requests.RefreshTokenRequest{})

			assert.NoError(t, err)
		})
	})

	t.Run("Given the refresh token does not parse", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRevokingAuthTestingSuite) {
			suite.mockRefreshTokenManager.EXPECT().
				Parse("the-refresh-token").
				Return(nil, errors.New("token is expired"))
		}

		t.Run("Should succeed without revoking", func(t *testing.T) {
			t.Parallel()

			suite := WhenRevokingAuthBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given a valid refresh token", func(t *testing.T) {
		t.Parallel()

		claims := &gateway_interfaces.RefreshTokenClaims{
			Address:          common_tests.RandomEthAddress(),
			SessionId:        "session-1",
			TokenId:          "token-1",
			SessionExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		initSuite := func(suite *WhenRevokingAuthTestingSuite) {
			suite.mockRefreshTokenManager.EXPECT().Parse("the-refresh-token").Return(claims, nil)
		}

		t.Run("Should revoke its session", func(t *testing.T) {
			t.Parallel()

			suite := WhenRevokingAuthBeforeEach(t)
			initSuite(suite)
			suite.mockRefreshSessionStore.EXPECT().Revoke("session-1", claims.SessionExpiresAt).Times(1)

			err := suite.sut.Execute(request)

			assert.NoError(t, err)
		})
	})
}
