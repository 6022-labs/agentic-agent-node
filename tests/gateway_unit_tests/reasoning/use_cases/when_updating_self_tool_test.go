package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingSelfToolTestingSuite struct {
	sut *use_cases.UpdateSelfTool

	mockClient *interfaces_mocks.MockSelfToolConfigClientInterface
}

func WhenUpdatingSelfToolBeforeEach(t *testing.T) *WhenUpdatingSelfToolTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockSelfToolConfigClientInterface(mockController)

	sut := use_cases.NewUpdateSelfTool(mockClient)

	return &WhenUpdatingSelfToolTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenUpdatingSelfTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the toggle", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			suite.mockClient.EXPECT().
				Update(gomock.Any(), id, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ uuid.UUID, body *models.ExternalUpdateSelfToolPayload) (*models.ExternalSelfToolPayload, error) {
					assert.False(t, body.Enabled)
					return &models.ExternalSelfToolPayload{
						Id:      id.String(),
						Name:    "run_command",
						Enabled: body.Enabled,
					}, nil
				})
		}

		t.Run("Should forward the flag and return the updated tool", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdateSelfToolRequest{Id: id, Enabled: lo.ToPtr(false)},
			)

			assert.NoError(t, err)
			assert.Equal(t, id.String(), resp.Id)
			assert.False(t, resp.Enabled)
		})
	})

	t.Run("Given a request without the enabled field", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateSelfToolRequest{Id: uuid.New()})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "enabled", validationErr.Field)
		})
	})

	t.Run("Given the runtime does not know the tool", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			suite.mockClient.EXPECT().
				Update(gomock.Any(), id, gomock.Any()).
				Return(nil, gateway_interfaces.ErrSelfToolNotFound)
		}

		t.Run("Should return a NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdateSelfToolRequest{Id: id, Enabled: lo.ToPtr(true)},
			)

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "self_tool_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the client fails with another error", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		clientErr := errors.New("runtime down")
		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			suite.mockClient.EXPECT().Update(gomock.Any(), id, gomock.Any()).Return(nil, clientErr)
		}

		t.Run("Should surface the client error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdateSelfToolRequest{Id: id, Enabled: lo.ToPtr(true)},
			)

			assert.ErrorIs(t, err, clientErr)
		})
	})
}
