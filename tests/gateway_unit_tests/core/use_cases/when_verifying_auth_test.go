package use_cases_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	common_mocks "github.com/6022-labs/agentic-agent-node/tests/common_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	gateway_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/core_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenVerifyingAuthTestingSuite struct {
	sut *use_cases.VerifyAuth

	mockNonceStore          *gateway_mocks.MockNonceStoreInterface
	mockSiweMessageBuilder  *gateway_mocks.MockSiweMessageBuilderInterface
	mockSignatureVerifier   *gateway_mocks.MockSignatureVerifierInterface
	mockAuthorizer          *common_mocks.MockAuthorizerInterface
	mockJwtIssuer           *gateway_mocks.MockJwtIssuerInterface
	mockRefreshTokenManager *gateway_mocks.MockRefreshTokenManagerInterface
}

func WhenVerifyingAuthBeforeEach(t *testing.T) *WhenVerifyingAuthTestingSuite {
	mockController := gomock.NewController(t)

	mockNonceStore := gateway_mocks.NewMockNonceStoreInterface(mockController)
	mockSiweMessageBuilder := gateway_mocks.NewMockSiweMessageBuilderInterface(mockController)
	mockSignatureVerifier := gateway_mocks.NewMockSignatureVerifierInterface(mockController)
	mockAuthorizer := common_mocks.NewMockAuthorizerInterface(mockController)
	mockJwtIssuer := gateway_mocks.NewMockJwtIssuerInterface(mockController)
	mockRefreshTokenManager := gateway_mocks.NewMockRefreshTokenManagerInterface(mockController)

	sut := use_cases.NewVerifyAuth(
		mockNonceStore,
		mockSiweMessageBuilder,
		mockSignatureVerifier,
		mockAuthorizer,
		mockJwtIssuer,
		mockRefreshTokenManager,
	)

	return &WhenVerifyingAuthTestingSuite{
		sut: sut,

		mockNonceStore:          mockNonceStore,
		mockSiweMessageBuilder:  mockSiweMessageBuilder,
		mockSignatureVerifier:   mockSignatureVerifier,
		mockAuthorizer:          mockAuthorizer,
		mockJwtIssuer:           mockJwtIssuer,
		mockRefreshTokenManager: mockRefreshTokenManager,
	}
}

var testSignature = "0x" + strings.Repeat("ab", 65)

func generateValidVerifyAuthRequest(address common.Address) *requests.VerifyAuthRequest {
	return &requests.VerifyAuthRequest{
		WalletAddressRequest: requests.WalletAddressRequest{Address: address.Hex()},
		Signature:            testSignature,
	}
}

func TestWhenVerifyingAuth(t *testing.T) {
	t.Parallel()

	address := common_tests.RandomEthAddress()
	entry := &gateway_interfaces.ChallengeEntry{
		Nonce:     "nonce-xyz",
		ChainId:   80002,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	siweMessage := "example.com wants you to sign in with your Ethereum account: ..."
	expectSiweMessage := func(suite *WhenVerifyingAuthTestingSuite) {
		suite.mockSiweMessageBuilder.EXPECT().
			Build(address, uint64(80002), "nonce-xyz", entry.IssuedAt, entry.ExpiresAt).
			Return(siweMessage)
	}

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.VerifyAuthRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the challenge is missing or expired", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume(address).Return(nil)
		}

		t.Run("Should return Unauthorized with challenge_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "challenge_not_found", unauthorizedErr.Code)
		})
	})

	t.Run("Given the signature verifier fails", func(t *testing.T) {
		t.Parallel()

		dependencyErr := errors.New("malformed")

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume(address).Return(entry)
			expectSiweMessage(suite)
			suite.mockSignatureVerifier.EXPECT().
				Recover(siweMessage, testSignature).
				Return(common.Address{}, dependencyErr)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, apperrors.InternalErrorMessage, internalErr.Message)
			assert.ErrorIs(t, err, dependencyErr)
		})
	})

	t.Run("Given the recovered signer differs from the claimed address", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume(address).Return(entry)
			expectSiweMessage(suite)
			suite.mockSignatureVerifier.EXPECT().
				Recover(siweMessage, testSignature).
				Return(common_tests.RandomEthAddress(), nil)
		}

		t.Run("Should return Unauthorized with signature_mismatch", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "signature_mismatch", unauthorizedErr.Code)
		})
	})

	t.Run("Given the authorizer reports identity-not-found", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume(address).Return(entry)
			expectSiweMessage(suite)
			suite.mockSignatureVerifier.EXPECT().Recover(siweMessage, testSignature).Return(address, nil)
			suite.mockAuthorizer.EXPECT().
				IsAuthorized(gomock.Any(), address).
				Return(false, fmt.Errorf("%w: https://rpc.example/secret-key", apperrors.ErrIdentityNotFound))
		}

		t.Run("Should return Unavailable with identity_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "identity_not_found", unavailableErr.Code)
			assert.Equal(t, apperrors.ErrIdentityNotFound.Error(), unavailableErr.Message)
			assert.ErrorIs(t, err, apperrors.ErrIdentityNotFound)
		})
	})

	t.Run("Given the authorizer fails", func(t *testing.T) {
		t.Parallel()

		dependencyErr := errors.New("rpc down")

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume(address).Return(entry)
			expectSiweMessage(suite)
			suite.mockSignatureVerifier.EXPECT().Recover(siweMessage, testSignature).Return(address, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(false, dependencyErr)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, apperrors.InternalErrorMessage, internalErr.Message)
			assert.ErrorIs(t, err, dependencyErr)
		})
	})

	t.Run("Given the authorizer rejects the address", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume(address).Return(entry)
			expectSiweMessage(suite)
			suite.mockSignatureVerifier.EXPECT().Recover(siweMessage, testSignature).Return(address, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(false, nil)
		}

		t.Run("Should return Unauthorized with not_authorized", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "not_authorized", unauthorizedErr.Code)
		})
	})

	t.Run("Given the refresh token manager fails", func(t *testing.T) {
		t.Parallel()

		dependencyErr := errors.New("entropy exhausted")

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume(address).Return(entry)
			expectSiweMessage(suite)
			suite.mockSignatureVerifier.EXPECT().Recover(siweMessage, testSignature).Return(address, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(true, nil)
			suite.mockRefreshTokenManager.EXPECT().Issue(address).Return(nil, dependencyErr)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "verify_failed", internalErr.Code)
			assert.Equal(t, apperrors.InternalErrorMessage, internalErr.Message)
			assert.ErrorIs(t, err, dependencyErr)
		})
	})

	t.Run("Given the jwt issuer fails", func(t *testing.T) {
		t.Parallel()

		dependencyErr := errors.New("signing failed")
		refreshToken := &gateway_interfaces.RefreshToken{
			Token:            "the-refresh-token",
			SessionId:        "session-1",
			SessionExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume(address).Return(entry)
			expectSiweMessage(suite)
			suite.mockSignatureVerifier.EXPECT().Recover(siweMessage, testSignature).Return(address, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(true, nil)
			suite.mockRefreshTokenManager.EXPECT().Issue(address).Return(refreshToken, nil)
			suite.mockJwtIssuer.EXPECT().
				Issue(address, "session-1", refreshToken.SessionExpiresAt).
				Return("", time.Time{}, dependencyErr)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "verify_failed", internalErr.Code)
			assert.ErrorIs(t, err, dependencyErr)
		})
	})

	t.Run("Given a valid nonce, signature, and authorized address", func(t *testing.T) {
		t.Parallel()

		expectedExp := time.Now().Add(30 * time.Minute)
		refreshToken := &gateway_interfaces.RefreshToken{
			Token:            "the-refresh-token",
			ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
			SessionId:        "session-1",
			SessionExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume(address).Return(entry)
			expectSiweMessage(suite)
			suite.mockSignatureVerifier.EXPECT().Recover(siweMessage, testSignature).Return(address, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(true, nil)
			suite.mockRefreshTokenManager.EXPECT().Issue(address).Return(refreshToken, nil)
			suite.mockJwtIssuer.EXPECT().
				Issue(address, "session-1", refreshToken.SessionExpiresAt).
				Return("the-token", expectedExp, nil)
		}

		t.Run("Should return Ok with a JWT response", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, "the-token", resp.Token)
			assert.Equal(t, expectedExp, resp.ExpiresAt)
		})

		t.Run("Should return the refresh token", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, refreshToken.Token, resp.RefreshToken)
			assert.Equal(t, refreshToken.ExpiresAt, resp.RefreshExpiresAt)
		})
	})
}
