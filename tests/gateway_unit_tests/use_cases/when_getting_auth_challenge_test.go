package use_cases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	gateway_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWhenGettingAuthChallenge(t *testing.T) {
	t.Parallel()

	t.Run("Given the nonce store returns a nonce + expiry", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nonceStore := gateway_mocks.NewMockNonceStoreInterface(ctrl)
		expectedExp := time.Now().Add(5 * time.Minute)
		nonceStore.EXPECT().Generate().Return("nonce-abc", expectedExp, nil)

		uc := use_cases.NewGetAuthChallenge(nonceStore)

		t.Run("Should return them in the response", func(t *testing.T) {
			resp, err := uc.Execute()
			assert.NoError(t, err)
			assert.Equal(t, "nonce-abc", resp.Nonce)
			assert.Equal(t, expectedExp, resp.ExpiresAt)
		})
	})

	t.Run("Given the nonce store fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nonceStore := gateway_mocks.NewMockNonceStoreInterface(ctrl)
		nonceStore.EXPECT().Generate().Return("", time.Time{}, errors.New("rng failed"))

		uc := use_cases.NewGetAuthChallenge(nonceStore)

		t.Run("Should propagate the error", func(t *testing.T) {
			_, err := uc.Execute()
			assert.ErrorContains(t, err, "rng failed")
		})
	})
}
