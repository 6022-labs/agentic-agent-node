package services_test

import (
	"context"
	"fmt"
	"testing"

	reasoning_services "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingMCPHeaderAuthSecretTestingSuite struct {
	sut *reasoning_services.MCPHeaderAuthSecretStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenDeletingMCPHeaderAuthSecretBeforeEach(t *testing.T) *WhenDeletingMCPHeaderAuthSecretTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, key string) (string, error) { return "scope/" + key, nil }).AnyTimes()

	return &WhenDeletingMCPHeaderAuthSecretTestingSuite{
		sut:         reasoning_services.NewMCPHeaderAuthSecretStorage(mockStorage, mockAgentScopedKeyBuilder),
		mockStorage: mockStorage,
	}
}

func TestWhenDeletingMCPHeaderAuthSecret(t *testing.T) {
	t.Parallel()

	t.Run("Given DeleteValue is called", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingMCPHeaderAuthSecretTestingSuite, connectionId uuid.UUID) {
			scopedKey := fmt.Sprintf("scope/mcp_auth_header_secret/%s/Authorization", connectionId)
			suite.mockStorage.EXPECT().Delete(gomock.Any(), scopedKey).Return(nil)
		}

		t.Run("Should delete the scoped key", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPHeaderAuthSecretBeforeEach(t)
			connectionId := uuid.New()
			initSuite(suite, connectionId)

			err := suite.sut.DeleteValue(context.Background(), connectionId, "Authorization")

			assert.NoError(t, err)
		})
	})
}
