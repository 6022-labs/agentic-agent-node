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

type WhenSavingLLMSecretTestingSuite struct {
	sut *reasoning_services.LLMSecretStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenSavingLLMSecretBeforeEach(t *testing.T) *WhenSavingLLMSecretTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, key string) (string, error) { return "scope/" + key, nil }).AnyTimes()

	return &WhenSavingLLMSecretTestingSuite{
		sut:         reasoning_services.NewLLMSecretStorage(mockStorage, mockAgentScopedKeyBuilder),
		mockStorage: mockStorage,
	}
}

func TestWhenSavingLLMSecret(t *testing.T) {
	t.Parallel()

	const scopedKey = "scope/llm_secret"

	t.Run("Given a blank api key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject without writing", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingLLMSecretBeforeEach(t)

			err := suite.sut.Save(context.Background(), "")

			assert.ErrorContains(t, err, "apiKey cannot be empty")
		})
	})

	t.Run("Given a valid api key", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingLLMSecretTestingSuite) {
			suite.mockStorage.EXPECT().
				Write(gomock.Any(), scopedKey, map[string]string{"apiKey": "sk-secret"}).
				Return(nil)
		}

		t.Run("Should write it to the scoped key", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingLLMSecretBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Save(context.Background(), "sk-secret")

			assert.NoError(t, err)
		})
	})
}
