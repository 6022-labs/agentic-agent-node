package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	common_mocks "github.com/6022-labs/agentic-agent-node/tests/common_mocks/interfaces_mocks"
	gateway_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenVerifyingAuthTestingSuite struct {
	sut *use_cases.VerifyAuth

	mockNonceStore        *gateway_mocks.MockNonceStoreInterface
	mockSignatureVerifier *gateway_mocks.MockSignatureVerifierInterface
	mockAuthorizer        *common_mocks.MockAuthorizerInterface
	mockJwtIssuer         *gateway_mocks.MockJwtIssuerInterface
}

func WhenVerifyingAuthBeforeEach(t *testing.T) *WhenVerifyingAuthTestingSuite {
	mockController := gomock.NewController(t)

	mockNonceStore := gateway_mocks.NewMockNonceStoreInterface(mockController)
	mockSignatureVerifier := gateway_mocks.NewMockSignatureVerifierInterface(mockController)
	mockAuthorizer := common_mocks.NewMockAuthorizerInterface(mockController)
	mockJwtIssuer := gateway_mocks.NewMockJwtIssuerInterface(mockController)

	sut := use_cases.NewVerifyAuth(mockNonceStore, mockSignatureVerifier, mockAuthorizer, mockJwtIssuer)

	return &WhenVerifyingAuthTestingSuite{
		sut: sut,

		mockNonceStore:        mockNonceStore,
		mockSignatureVerifier: mockSignatureVerifier,
		mockAuthorizer:        mockAuthorizer,
		mockJwtIssuer:         mockJwtIssuer,
	}
}

func generateValidVerifyAuthRequest(address common.Address) *requests.VerifyAuthRequest {
	return &requests.VerifyAuthRequest{
		Address:   address.Hex(),
		Signature: "0xsig",
		Nonce:     "nonce-xyz",
	}
}

func TestWhenVerifyingAuth(t *testing.T) {
	t.Parallel()

	address := common.HexToAddress("0x3333333333333333333333333333333333333333")

	t.Run("Given the nonce is missing or expired", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume("nonce-xyz").Return(false)
		}

		t.Run("Should return ErrNonceNotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.ErrorIs(t, err, use_cases.ErrNonceNotFound)
		})
	})

	t.Run("Given the recovered signer differs from the claimed address", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume("nonce-xyz").Return(true)
			suite.mockSignatureVerifier.EXPECT().Recover("nonce-xyz", "0xsig").Return(common.HexToAddress("0xdeadbeef"), nil)
		}

		t.Run("Should return ErrSignatureMismatch", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.ErrorIs(t, err, use_cases.ErrSignatureMismatch)
		})
	})

	t.Run("Given the signature verifier fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume("nonce-xyz").Return(true)
			suite.mockSignatureVerifier.EXPECT().Recover("nonce-xyz", "0xsig").Return(common.Address{}, errors.New("malformed"))
		}

		t.Run("Should propagate the verifier error", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.ErrorContains(t, err, "malformed")
		})
	})

	t.Run("Given the authorizer rejects the address", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume("nonce-xyz").Return(true)
			suite.mockSignatureVerifier.EXPECT().Recover("nonce-xyz", "0xsig").Return(address, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(false, nil)
		}

		t.Run("Should return ErrAddressNotAuthorized", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.ErrorIs(t, err, use_cases.ErrAddressNotAuthorized)
		})
	})

	t.Run("Given the authorizer errors", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume("nonce-xyz").Return(true)
			suite.mockSignatureVerifier.EXPECT().Recover("nonce-xyz", "0xsig").Return(address, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(false, errors.New("rpc down"))
		}

		t.Run("Should propagate the authorizer error", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			assert.ErrorContains(t, err, "rpc down")
		})
	})

	t.Run("Given a valid nonce, signature, and authorized address", func(t *testing.T) {
		t.Parallel()

		expectedExp := time.Now().Add(30 * time.Minute)
		initSuite := func(suite *WhenVerifyingAuthTestingSuite) {
			suite.mockNonceStore.EXPECT().Consume("nonce-xyz").Return(true)
			suite.mockSignatureVerifier.EXPECT().Recover("nonce-xyz", "0xsig").Return(address, nil)
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(true, nil)
			suite.mockJwtIssuer.EXPECT().Issue(address).Return("the-token", expectedExp, nil)
		}

		t.Run("Should return a JWT response", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAuthBeforeEach(t)
			request := generateValidVerifyAuthRequest(address)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, "the-token", resp.Token)
			assert.Equal(t, expectedExp, resp.ExpiresAt)
		})
	})
}
