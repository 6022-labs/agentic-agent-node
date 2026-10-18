package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentMintInfoTestingSuite struct {
	sut *use_cases.GetAgentMintInfo

	mockAgentRepository          *repository_mocks.MockAgentRepositoryInterface
	mockAgentAddressRepository   *repository_mocks.MockAgentAddressRepositoryInterface
	mockAgentCollectionRequester *interface_mocks.MockAgentCollectionRequesterInterface
}

func WhenGettingAgentMintInfoBeforeEach(t *testing.T) *WhenGettingAgentMintInfoTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentAddressRepository := repository_mocks.NewMockAgentAddressRepositoryInterface(mockController)
	mockAgentCollectionRequester := interface_mocks.NewMockAgentCollectionRequesterInterface(mockController)

	sut := use_cases.NewGetAgentMintInfo(mockAgentRepository, mockAgentAddressRepository, mockAgentCollectionRequester)

	return &WhenGettingAgentMintInfoTestingSuite{
		sut:                          sut,
		mockAgentRepository:          mockAgentRepository,
		mockAgentAddressRepository:   mockAgentAddressRepository,
		mockAgentCollectionRequester: mockAgentCollectionRequester,
	}
}

func TestWhenGettingAgentMintInfo(t *testing.T) {
	t.Parallel()

	owner := common_tests.RandomEthAddress().Hex()

	t.Run("Given no registration draft exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentMintInfoTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return a NotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentMintInfoBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given a draft on a non-moderated collection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentMintInfoTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithOwner(owner).WithName("alice").Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[enums.AddressType]string{enums.AddressTypeEvm: "0xabc"}, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				ModeratorCount(gomock.Any(), gomock.Any()).
				Return(uint64(0), nil)
		}

		t.Run("Should assemble the calldata and report no moderators", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentMintInfoBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "alice", resp.Name)
			assert.Equal(t, uint64(0), resp.ModeratorCount)
			assert.False(t, resp.OwnerIsModerator)
		})
	})

	t.Run("Given a draft on a moderated collection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentMintInfoTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithOwner(owner).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[enums.AddressType]string{enums.AddressTypeEvm: "0xabc"}, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				ModeratorCount(gomock.Any(), gomock.Any()).
				Return(uint64(2), nil)
			suite.mockAgentCollectionRequester.EXPECT().
				IsModerator(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(true, nil)
		}

		t.Run("Should report the owner's moderator status", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentMintInfoBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, uint64(2), resp.ModeratorCount)
			assert.True(t, resp.OwnerIsModerator)
		})
	})
}
