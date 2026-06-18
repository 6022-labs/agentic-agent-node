package services_test

import (
	"context"
	"testing"

	reasoning_services "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingLLMSecretTestingSuite struct {
	sut *reasoning_services.LLMSecretStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenDeletingLLMSecretBeforeEach(t *testing.T) *WhenDeletingLLMSecretTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, key string) (string, error) { return "scope/" + key, nil }).AnyTimes()

	return &WhenDeletingLLMSecretTestingSuite{
		sut:         reasoning_services.NewLLMSecretStorage(mockStorage, mockAgentScopedKeyBuilder),
		mockStorage: mockStorage,
	}
}

func TestWhenDeletingLLMSecret(t *testing.T) {
	t.Parallel()

	const scopedKey = "scope/llm_secret"

	t.Run("Given Delete is called", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMSecretTestingSuite) {
			suite.mockStorage.EXPECT().Delete(gomock.Any(), scopedKey).Return(nil)
		}

		t.Run("Should delete the scoped key", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMSecretBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background())

			assert.NoError(t, err)
		})
	})
}
