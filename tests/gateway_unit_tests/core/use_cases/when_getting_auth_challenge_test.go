package use_cases_test

import (
	"errors"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	gateway_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/core_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const challengeWalletHex = "0xd389d997Df5654b87914492E1e0C6861bfe835DE"

type WhenGettingAuthChallengeTestingSuite struct {
	sut *use_cases.GetAuthChallenge

	mockNonceStore         *gateway_mocks.MockNonceStoreInterface
	mockSiweMessageBuilder *gateway_mocks.MockSiweMessageBuilderInterface
}

func WhenGettingAuthChallengeBeforeEach(t *testing.T) *WhenGettingAuthChallengeTestingSuite {
	mockController := gomock.NewController(t)

	mockNonceStore := gateway_mocks.NewMockNonceStoreInterface(mockController)
	mockSiweMessageBuilder := gateway_mocks.NewMockSiweMessageBuilderInterface(mockController)

	sut := use_cases.NewGetAuthChallenge(mockNonceStore, mockSiweMessageBuilder)

	return &WhenGettingAuthChallengeTestingSuite{
		sut: sut,

		mockNonceStore:         mockNonceStore,
		mockSiweMessageBuilder: mockSiweMessageBuilder,
	}
}

func TestWhenGettingAuthChallenge(t *testing.T) {
	t.Parallel()

	wallet := common.HexToAddress(challengeWalletHex)

	t.Run("Given the nonce store issues a challenge", func(t *testing.T) {
		t.Parallel()

		entry := &gateway_interfaces.ChallengeEntry{
			Nonce:     "nonce-abc",
			ChainId:   80002,
			IssuedAt:  time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}
		initSuite := func(suite *WhenGettingAuthChallengeTestingSuite) {
			suite.mockNonceStore.EXPECT().Generate(wallet, uint64(80002)).Return(entry, nil)
			suite.mockSiweMessageBuilder.EXPECT().
				Build(wallet, uint64(80002), "nonce-abc", entry.IssuedAt, entry.ExpiresAt).
				Return("siwe message")
		}

		t.Run("Should return the message, nonce and expiry", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAuthChallengeBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(&requests.GetAuthChallengeRequest{Address: challengeWalletHex, ChainId: 80002})

			assert.NoError(t, err)
			assert.Equal(t, "siwe message", resp.Message)
			assert.Equal(t, entry.ExpiresAt, resp.ExpiresAt)
		})
	})

	t.Run("Given an invalid address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without generating a nonce", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAuthChallengeBeforeEach(t)

			_, err := suite.sut.Execute(&requests.GetAuthChallengeRequest{Address: "not-an-address"})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "address", validationErr.Field)
		})
	})

	t.Run("Given the nonce store fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAuthChallengeTestingSuite) {
			suite.mockNonceStore.EXPECT().Generate(wallet, uint64(80002)).Return(nil, errors.New("rng failed"))
		}

		t.Run("Should return InternalError carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAuthChallengeBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(&requests.GetAuthChallengeRequest{Address: challengeWalletHex, ChainId: 80002})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "rng failed")
		})
	})
}
