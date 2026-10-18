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

type WhenSavingMCPHeaderAuthSecretTestingSuite struct {
	sut *reasoning_services.MCPHeaderAuthSecretStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenSavingMCPHeaderAuthSecretBeforeEach(t *testing.T) *WhenSavingMCPHeaderAuthSecretTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, key string) (string, error) { return "scope/" + key, nil }).AnyTimes()

	return &WhenSavingMCPHeaderAuthSecretTestingSuite{
		sut:         reasoning_services.NewMCPHeaderAuthSecretStorage(mockStorage, mockAgentScopedKeyBuilder),
		mockStorage: mockStorage,
	}
}

func TestWhenSavingMCPHeaderAuthSecret(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank header name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject without writing", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingMCPHeaderAuthSecretBeforeEach(t)

			err := suite.sut.SaveValue(context.Background(), uuid.New(), "", "value")

			assert.ErrorContains(t, err, "header key cannot be empty")
		})
	})

	t.Run("Given a blank header value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject without writing", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingMCPHeaderAuthSecretBeforeEach(t)

			err := suite.sut.SaveValue(context.Background(), uuid.New(), "Authorization", "")

			assert.ErrorContains(t, err, "header value cannot be empty")
		})
	})

	t.Run("Given two connections share a header name", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingMCPHeaderAuthSecretTestingSuite, a, b uuid.UUID) {
			keyA := fmt.Sprintf("scope/mcp_auth_header_secret/%s/Authorization", a)
			keyB := fmt.Sprintf("scope/mcp_auth_header_secret/%s/Authorization", b)
			suite.mockStorage.EXPECT().Write(gomock.Any(), keyA, map[string]string{"value": "Bearer aaa"}).Return(nil)
			suite.mockStorage.EXPECT().Write(gomock.Any(), keyB, map[string]string{"value": "Bearer bbb"}).Return(nil)
		}

		t.Run("Should write each value under a key scoped by connectionId", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingMCPHeaderAuthSecretBeforeEach(t)
			a := uuid.New()
			b := uuid.New()
			initSuite(suite, a, b)

			assert.NoError(t, suite.sut.SaveValue(context.Background(), a, "Authorization", "Bearer aaa"))
			assert.NoError(t, suite.sut.SaveValue(context.Background(), b, "Authorization", "Bearer bbb"))
		})
	})
}
