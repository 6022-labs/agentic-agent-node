package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingAgentCollectionsTestingSuite struct {
	sut *use_cases.ListAgentCollections

	mockAgentCollectionsManagerRequester *interface_mocks.MockAgentCollectionsManagerRequesterInterface
}

func WhenListingAgentCollectionsBeforeEach(t *testing.T) *WhenListingAgentCollectionsTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentCollectionsManagerRequester := interface_mocks.NewMockAgentCollectionsManagerRequesterInterface(
		mockController,
	)

	sut := use_cases.NewListAgentCollections(mockAgentCollectionsManagerRequester)

	return &WhenListingAgentCollectionsTestingSuite{
		sut: sut,

		mockAgentCollectionsManagerRequester: mockAgentCollectionsManagerRequester,
	}
}

func TestWhenListingAgentCollections(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.ListAgentCollectionsRequest{ChainIdRequest: requests.ChainIdRequest{ChainId: "  "}},
			)

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the chain read fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockAgentCollectionsManagerRequester.EXPECT().
				ListCollections(gomock.Any()).
				Return(nil, errors.New("rpc down"))
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.ListAgentCollectionsRequest{ChainIdRequest: requests.ChainIdRequest{ChainId: "80002"}},
			)

			var target *apperrors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the manager governs no collections", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockAgentCollectionsManagerRequester.EXPECT().
				ListCollections(gomock.Any()).
				Return([]models.AgentCollection{}, nil)
		}

		t.Run("Should return an empty list, not null", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(
				context.Background(),
				&requests.ListAgentCollectionsRequest{ChainIdRequest: requests.ChainIdRequest{ChainId: "80002"}},
			)

			assert.NoError(t, err)
			assert.NotNil(t, resp.Collections)
			assert.Empty(t, resp.Collections)
		})
	})

	t.Run("Given the manager governs two collections", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockAgentCollectionsManagerRequester.EXPECT().
				ListCollections(gomock.Any()).
				Return([]models.AgentCollection{
					{Address: "0xaaa", Name: "alpha", Description: "first", ModeratorCount: 2, NextTokenId: 7},
					{Address: "0xbbb", Name: "beta", Description: "second", ModeratorCount: 0, NextTokenId: 1},
				}, nil)
		}

		t.Run("Should return them in order with their moderator state", func(t *testing.T) {
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
			assert.Equal(t, uint64(1), resp.Collections[1].NextTokenId)
		})
	})
}
