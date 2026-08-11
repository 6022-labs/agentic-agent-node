package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingSelfToolsTestingSuite struct {
	sut *use_cases.ListSelfTools

	mockClient *interfaces_mocks.MockSelfToolConfigClientInterface
}

func WhenListingSelfToolsBeforeEach(t *testing.T) *WhenListingSelfToolsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockSelfToolConfigClientInterface(mockController)

	sut := use_cases.NewListSelfTools(mockClient)

	return &WhenListingSelfToolsTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenListingSelfTools(t *testing.T) {
	t.Parallel()

	t.Run("Given the client returns two tools", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingSelfToolsTestingSuite) {
			suite.mockClient.EXPECT().List(gomock.Any()).Return(&models.ExternalSelfToolListPayload{
				Tools: []models.ExternalSelfToolPayload{
					{Id: "a", Name: "run_command", Description: "runs a command", Enabled: true},
					{Id: "b", Name: "list_processes", Description: "lists processes", Enabled: false},
				},
			}, nil)
		}

		t.Run("Should return Ok with the rows on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Len(t, resp.Tools, 2)
			assert.Equal(t, "run_command", resp.Tools[0].Name)
			assert.True(t, resp.Tools[0].Enabled)
			assert.False(t, resp.Tools[1].Enabled)
		})
	})

	t.Run("Given the client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingSelfToolsTestingSuite) {
			suite.mockClient.EXPECT().List(gomock.Any()).Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})
}
