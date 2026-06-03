package use_cases_test

import (
	"context"
	"errors"
	"testing"

	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentIdentityTestingSuite struct {
	sut *use_cases.GetAgentIdentity

	mockAgentRepository *repository_mocks.MockAgentRepositoryInterface
	mockStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
}

func WhenGettingAgentIdentityBeforeEach(t *testing.T) *WhenGettingAgentIdentityTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(mockController)

	sut := use_cases.NewGetAgentIdentity(mockAgentRepository, mockStateRepository)

	return &WhenGettingAgentIdentityTestingSuite{
		sut:                 sut,
		mockAgentRepository: mockAgentRepository,
		mockStateRepository: mockStateRepository,
	}
}

func TestWhenGettingAgentIdentity(t *testing.T) {
	t.Parallel()

	t.Run("Given the mint is confirmed and the agent is persisted", func(t *testing.T) {
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
			suite.mockStateRepository.EXPECT().Get(gomock.Any()).Return(state, nil)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
		}

		t.Run("Should return Ok with the identity on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.GetAgentIdentityResponse)
			assert.True(t, ok)
			assert.Equal(t, "0xcoll", body.CollectionAddress)
			assert.Equal(t, uint64(42), body.TokenId)
			assert.Equal(t, "Alice", body.Name)
		})
	})

	t.Run("Given no registration state is persisted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentIdentityTestingSuite) {
			suite.mockStateRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return NotFound with identity_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeNotFound, res.Outcome)
			assert.Equal(t, "identity_not_found", res.Code)
		})
	})

	t.Run("Given the mint is not yet confirmed", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentIdentityTestingSuite) {
			state := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusSubmitted).
				Build()
			suite.mockStateRepository.EXPECT().Get(gomock.Any()).Return(state, nil)
		}

		t.Run("Should return NotFound with identity_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeNotFound, res.Outcome)
			assert.Equal(t, "identity_not_found", res.Code)
		})
	})

	t.Run("Given the state read fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentIdentityTestingSuite) {
			suite.mockStateRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should return InternalError carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeInternalError, res.Outcome)
			assert.Contains(t, res.Message, "db down")
		})
	})
}
