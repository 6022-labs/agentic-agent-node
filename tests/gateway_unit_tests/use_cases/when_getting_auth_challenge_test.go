package use_cases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	gateway_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAuthChallengeTestingSuite struct {
	sut *use_cases.GetAuthChallenge

	mockNonceStore *gateway_mocks.MockNonceStoreInterface
}

func WhenGettingAuthChallengeBeforeEach(t *testing.T) *WhenGettingAuthChallengeTestingSuite {
	mockController := gomock.NewController(t)

	mockNonceStore := gateway_mocks.NewMockNonceStoreInterface(mockController)

	sut := use_cases.NewGetAuthChallenge(mockNonceStore)

	return &WhenGettingAuthChallengeTestingSuite{
		sut: sut,

		mockNonceStore: mockNonceStore,
	}
}

func TestWhenGettingAuthChallenge(t *testing.T) {
	t.Parallel()

	t.Run("Given the nonce store returns a nonce + expiry", func(t *testing.T) {
		t.Parallel()

		expectedExp := time.Now().Add(5 * time.Minute)
		initSuite := func(suite *WhenGettingAuthChallengeTestingSuite) {
			suite.mockNonceStore.EXPECT().Generate().Return("nonce-abc", expectedExp, nil)
		}

		t.Run("Should return Ok with the nonce in the response", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAuthChallengeBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute()

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.GetAuthChallengeResponse)
			assert.True(t, ok)
			assert.Equal(t, "nonce-abc", body.Nonce)
			assert.Equal(t, expectedExp, body.ExpiresAt)
		})
	})

	t.Run("Given the nonce store fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAuthChallengeTestingSuite) {
			suite.mockNonceStore.EXPECT().Generate().Return("", time.Time{}, errors.New("rng failed"))
		}

		t.Run("Should return InternalError carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAuthChallengeBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute()

			assert.Equal(t, usecase_result.OutcomeInternalError, res.Outcome)
			assert.Contains(t, res.Message, "rng failed")
		})
	})
}
