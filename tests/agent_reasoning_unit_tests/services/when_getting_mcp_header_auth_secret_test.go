package services_test

import (
	"context"
	"fmt"
	"testing"

	reasoning_services "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingMCPHeaderAuthSecretTestingSuite struct {
	sut *reasoning_services.MCPHeaderAuthSecretStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenGettingMCPHeaderAuthSecretBeforeEach(t *testing.T) *WhenGettingMCPHeaderAuthSecretTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, key string) (string, error) { return "scope/" + key, nil }).AnyTimes()

	return &WhenGettingMCPHeaderAuthSecretTestingSuite{
		sut:         reasoning_services.NewMCPHeaderAuthSecretStorage(mockStorage, mockAgentScopedKeyBuilder),
		mockStorage: mockStorage,
	}
}

func TestWhenGettingMCPHeaderAuthSecret(t *testing.T) {
	t.Parallel()

	t.Run("Given no header value has been saved", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingMCPHeaderAuthSecretTestingSuite, connectionId uuid.UUID) {
			scopedKey := fmt.Sprintf("scope/mcp_auth_header_secret/%s/Authorization", connectionId)
			suite.mockStorage.EXPECT().
				Read(gomock.Any(), scopedKey).
				Return(nil, runtime_common_interfaces.ErrSecretNotFound)
		}

		t.Run("Should report not-found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingMCPHeaderAuthSecretBeforeEach(t)
			connectionId := uuid.New()
			initSuite(suite, connectionId)

			value, found, err := suite.sut.GetValue(context.Background(), connectionId, "Authorization")

			assert.NoError(t, err)
			assert.False(t, found)
			assert.Empty(t, value)
		})
	})

	t.Run("Given a header value has been saved", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingMCPHeaderAuthSecretTestingSuite, connectionId uuid.UUID) {
			scopedKey := fmt.Sprintf("scope/mcp_auth_header_secret/%s/Authorization", connectionId)
			suite.mockStorage.EXPECT().
				Read(gomock.Any(), scopedKey).
				Return(map[string]string{"value": "Bearer abc"}, nil)
		}

		t.Run("Should return the header value", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingMCPHeaderAuthSecretBeforeEach(t)
			connectionId := uuid.New()
			initSuite(suite, connectionId)

			value, found, err := suite.sut.GetValue(context.Background(), connectionId, "Authorization")

			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, "Bearer abc", value)
		})
	})
}
