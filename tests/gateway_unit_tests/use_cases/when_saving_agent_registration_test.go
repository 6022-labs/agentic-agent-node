package use_cases_test

import (
	"context"
	"mime/multipart"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSavingAgentRegistrationTestingSuite struct {
	sut *use_cases.SaveAgentRegistration

	mockClient *interfaces_mocks.MockAgentRegistrationClientInterface
}

func WhenSavingAgentRegistrationBeforeEach(t *testing.T) *WhenSavingAgentRegistrationTestingSuite {
	mockController := gomock.NewController(t)
	mockClient := interfaces_mocks.NewMockAgentRegistrationClientInterface(mockController)
	sut := use_cases.NewSaveAgentRegistration(mockClient)
	return &WhenSavingAgentRegistrationTestingSuite{sut: sut, mockClient: mockClient}
}

func TestWhenSavingAgentRegistration(t *testing.T) {
	t.Parallel()

	const owner = "0x1111111111111111111111111111111111111111"

	t.Run("Given a request that fails validation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.SaveAgentRegistrationRequest{Name: "Alice"}, owner)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime accepts the draft", func(t *testing.T) {
		t.Parallel()

		var forwarded *models.ExternalSaveAgentRegistrationPayload
		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockClient.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, body *models.ExternalSaveAgentRegistrationPayload) (*models.ExternalAgentRegistrationPayload, error) {
					forwarded = body
					return &models.ExternalAgentRegistrationPayload{Name: body.Name}, nil
				})
		}

		t.Run("Should inject the authenticated owner into the forwarded body", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.SaveAgentRegistrationRequest{
				CollectionAddress: "0x2222222222222222222222222222222222222222",
				Name:              "Alice",
				Role:              "assistant",
				Image:             &multipart.FileHeader{Filename: "avatar.png", Size: 128},
			}, owner)

			assert.NoError(t, err)
			assert.Equal(t, "Alice", resp.Name)
			assert.Equal(t, owner, forwarded.Owner)
			assert.Equal(t, "Alice", forwarded.Name)
		})
	})

	t.Run("Given the client returns a domain error", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockClient.EXPECT().Save(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewNotFoundError("draft_not_found", "no draft"))
		}

		t.Run("Should propagate it unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.SaveAgentRegistrationRequest{
				CollectionAddress: "0x2222222222222222222222222222222222222222",
				Name:              "Alice",
				Role:              "assistant",
				Image:             &multipart.FileHeader{Filename: "avatar.png", Size: 128},
			}, owner)

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "draft_not_found", notFoundErr.Code)
		})
	})
}
