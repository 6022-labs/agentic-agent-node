package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/requests"
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

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockClient.EXPECT().List(gomock.Any(), "80002").
				Return(nil, errors.NewUnavailableError("runtime_unavailable", "runtime down", nil))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.ListAgentCollectionsRequest{ChainIdRequest: requests.ChainIdRequest{ChainId: "80002"}},
			)

			var unavailableErr *errors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockClient.EXPECT().List(gomock.Any(), "80002").
				Return(nil, errors.NewValidationError("chain_id", "chain id is invalid"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.ListAgentCollectionsRequest{ChainIdRequest: requests.ChainIdRequest{ChainId: "80002"}},
			)

			var validationErr *errors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "chain_id", validationErr.Field)
		})
	})

	t.Run("Given the runtime returns the collections catalog", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockClient.EXPECT().List(gomock.Any(), "80002").Return(&models.ExternalAgentCollectionsPayload{
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

			resp, err := suite.sut.Execute(
				context.Background(),
				&requests.ListAgentCollectionsRequest{ChainIdRequest: requests.ChainIdRequest{ChainId: "80002"}},
			)

			assert.NoError(t, err)
			assert.Len(t, resp.Collections, 2)
			assert.Equal(t, "alpha", resp.Collections[0].Name)
			assert.Equal(t, uint64(2), resp.Collections[0].ModeratorCount)
			assert.Equal(t, "0xbbb", resp.Collections[1].Address)
		})
	})
}
