package use_cases_test

import (
	"context"
	"errors"
	"testing"

	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentIdentityTestingSuite struct {
	sut *use_cases.GetAgentIdentity

	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
}

func WhenGettingAgentIdentityBeforeEach(t *testing.T) *WhenGettingAgentIdentityTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(
		mockController,
	)

	sut := use_cases.NewGetAgentIdentity(mockAgentRepository, mockAgentRegistrationStateRepository)

	return &WhenGettingAgentIdentityTestingSuite{
		sut:                                  sut,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
	}
}

func TestWhenGettingAgentIdentity(t *testing.T) {
	t.Parallel()

	t.Run("Given the registration state lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentIdentityTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(42).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "db down")
		})
	})

	t.Run("Given no registration state exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentIdentityTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(42).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, nil)
		}

		t.Run("Should return NotFound with identity_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "identity_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the mint is not yet confirmed", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentIdentityTestingSuite) {
			state := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusSubmitted).
				Build()
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(42).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(state, nil)
		}

		t.Run("Should return NotFound with identity_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "identity_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the mint is confirmed and the agent exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentIdentityTestingSuite) {
			state := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusConfirmed).
				Build()
			agent := model_builders.NewAgentBuilder().
				WithCollectionAddress("0xcoll").
				WithCollectionTokenId(42).
				WithName("Alice").
				Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(state, nil)
		}

		t.Run("Should return Ok with the identity on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "0xcoll", resp.CollectionAddress)
			assert.Equal(t, uint64(42), resp.TokenId)
			assert.Equal(t, "Alice", resp.Name)
		})
	})
}
