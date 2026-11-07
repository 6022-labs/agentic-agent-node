package use_cases_test

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingConversationTestingSuite struct {
	sut *use_cases.UpdateConversation

	mockClient *interfaces_mocks.MockConversationClientInterface
}

func WhenUpdatingConversationBeforeEach(t *testing.T) *WhenUpdatingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockConversationClientInterface(mockController)

	return &WhenUpdatingConversationTestingSuite{
		sut:        use_cases.NewUpdateConversation(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenUpdatingConversation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "  "},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Solidity audits"},
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				UpdateConversation(gomock.Any(), "c1", gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down", nil))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Solidity audits"},
			})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				UpdateConversation(gomock.Any(), "c1", gomock.Any()).
				Return(nil, apperrors.NewValidationError("name", "cannot be empty"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Solidity audits"},
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given the runtime receives no caller id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				UpdateConversation(gomock.Any(), "c1", gomock.Any()).
				Return(nil, apperrors.NewUnauthorizedError("missing_caller_id", "an authenticated caller id is required"))
		}

		t.Run("Should surface the UnauthorizedError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Solidity audits"},
			})

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the conversation is out of the caller's view", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				UpdateConversation(gomock.Any(), "ghost", gomock.Any()).
				Return(nil, apperrors.NewNotFoundError("conversation_not_found", "conversation not found"))
		}

		t.Run("Should surface the NotFoundError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "ghost"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Solidity audits"},
			})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "conversation_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given another conversation holds that name", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				UpdateConversation(gomock.Any(), "c1", gomock.Any()).
				Return(nil, apperrors.NewConflictError("conversation_name_taken", "already used"))
		}

		t.Run("Should surface the ConflictError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Solidity audits"},
			})

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "conversation_name_taken", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				UpdateConversation(gomock.Any(), "c1", gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down", nil))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Solidity audits"},
			})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime renames the session", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				UpdateConversation(gomock.Any(), "c1", gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					_ string,
					body *models.ExternalUpdateConversationPayload,
				) (*models.ExternalConversationPayload, error) {
					assert.Equal(t, "Solidity audits", body.Name)
					return &models.ExternalConversationPayload{
						ConversationId: "c1",
						Name:           lo.ToPtr("Solidity audits"),
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil
				})
		}

		t.Run("Should map the payload onto the wire response", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Solidity audits"},
			})

			assert.NoError(t, err)
			assert.Equal(t, "Solidity audits", *resp.Name)
			assert.Equal(t, now, resp.UpdatedAt)
		})
	})
}
