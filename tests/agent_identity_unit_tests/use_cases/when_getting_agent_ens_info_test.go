package use_cases_test

import (
	"context"
	"testing"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentEnsInfoTestingSuite struct {
	sut *use_cases.GetAgentEnsInfo

	mockAgentRepository        *repository_mocks.MockAgentRepositoryInterface
	mockRuntimeRecordsProvider *interface_mocks.MockRuntimeRecordsProviderInterface
}

func WhenGettingAgentEnsInfoBeforeEach(t *testing.T) *WhenGettingAgentEnsInfoTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockRuntimeRecordsProvider := interface_mocks.NewMockRuntimeRecordsProviderInterface(mockController)

	sut := use_cases.NewGetAgentEnsInfo(mockAgentRepository, mockRuntimeRecordsProvider)

	return &WhenGettingAgentEnsInfoTestingSuite{
		sut:                        sut,
		mockAgentRepository:        mockAgentRepository,
		mockRuntimeRecordsProvider: mockRuntimeRecordsProvider,
	}
}

func TestWhenGettingAgentEnsInfo(t *testing.T) {
	t.Parallel()

	t.Run("Given a minted agent", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentEnsInfoTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(7).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockRuntimeRecordsProvider.EXPECT().RuntimeRecords().Return([]identity_interfaces.EnsRecord{{Key: "url", Value: "https://agent.example.com"}})
		}

		t.Run("Should return the tokenId and records to publish", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentEnsInfoBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, uint64(7), resp.TokenId)
			assert.Len(t, resp.Records, 1)
		})
	})

	t.Run("Given no draft has been saved", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentEnsInfoTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return a NotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentEnsInfoBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentEnsInfoTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
		}

		t.Run("Should return a Conflict", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentEnsInfoBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
		})
	})
}
