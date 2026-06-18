package services_test

import (
	"context"
	"testing"

	reasoning_services "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingLLMSecretTestingSuite struct {
	sut *reasoning_services.LLMSecretStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenGettingLLMSecretBeforeEach(t *testing.T) *WhenGettingLLMSecretTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, key string) (string, error) { return "scope/" + key, nil }).AnyTimes()

	return &WhenGettingLLMSecretTestingSuite{
		sut:         reasoning_services.NewLLMSecretStorage(mockStorage, mockAgentScopedKeyBuilder),
		mockStorage: mockStorage,
	}
}

func TestWhenGettingLLMSecret(t *testing.T) {
	t.Parallel()

	const scopedKey = "scope/llm_secret"

	t.Run("Given no api key has been saved", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLLMSecretTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(nil, runtime_common_interfaces.ErrSecretNotFound)
		}

		t.Run("Should report not-found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMSecretBeforeEach(t)
			initSuite(suite)

			apiKey, found, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.False(t, found)
			assert.Empty(t, apiKey)
		})
	})

	t.Run("Given an api key has been saved", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLLMSecretTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(map[string]string{"apiKey": "sk-secret-123"}, nil)
		}

		t.Run("Should return the api key", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMSecretBeforeEach(t)
			initSuite(suite)

			apiKey, found, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, "sk-secret-123", apiKey)
		})
	})
}
