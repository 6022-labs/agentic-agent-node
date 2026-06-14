package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/identity_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingAgentCollectionsTestingSuite struct {
	sut *use_cases.ListAgentCollections

	mockClient *interfaces_mocks.MockAgentCollectionsClientInterface
}

func WhenListingAgentCollectionsBeforeEach(t *testing.T) *WhenListingAgentCollectionsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockAgentCollectionsClientInterface(mockController)

	sut := use_cases.NewListAgentCollections(mockClient)

	return &WhenListingAgentCollectionsTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenListingAgentCollections(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns two collections", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockClient.EXPECT().List(gomock.Any()).Return(&models.ExternalAgentCollectionsPayload{
				Collections: []models.ExternalAgentCollectionPayload{
					{Address: "0xaaa", Name: "alpha", Description: "first", ModeratorCount: 2, NextTokenId: 7},
					{Address: "0xbbb", Name: "beta", Description: "second", ModeratorCount: 0, NextTokenId: 1},
				},
			}, nil)
		}

		t.Run("Should return the rows on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Len(t, resp.Collections, 2)
			assert.Equal(t, "alpha", resp.Collections[0].Name)
			assert.Equal(t, uint64(2), resp.Collections[0].ModeratorCount)
			assert.Equal(t, "0xbbb", resp.Collections[1].Address)
		})
	})

	t.Run("Given the runtime is unreachable", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockClient.EXPECT().List(gomock.Any()).
				Return(nil, errors.NewUnavailableError("runtime_unavailable", "down"))
		}

		t.Run("Should propagate the UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var target *errors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})
}
