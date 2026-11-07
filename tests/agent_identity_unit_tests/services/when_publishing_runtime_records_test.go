package services_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const runtimeUrl = "https://agent.example.com"

var runtimeRecords = []identity_interfaces.EnsRecord{{Key: "url", Value: runtimeUrl}}

type WhenPublishingRuntimeRecordsTestingSuite struct {
	sut *services.EnsPublisher

	mockEnsRegistrar                     *interface_mocks.MockEnsRegistrarInterface
	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
	mockRuntimeRecordsProvider           *interface_mocks.MockRuntimeRecordsProviderInterface
}

func WhenPublishingRuntimeRecordsBeforeEach(t *testing.T) *WhenPublishingRuntimeRecordsTestingSuite {
	mockController := gomock.NewController(t)

	mockEnsRegistrar := interface_mocks.NewMockEnsRegistrarInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(
		mockController,
	)
	mockRuntimeRecordsProvider := interface_mocks.NewMockRuntimeRecordsProviderInterface(mockController)

	sut := services.NewEnsPublisher(
		zap.NewNop(),
		mockEnsRegistrar,
		mockAgentRepository,
		mockRuntimeRecordsProvider,
		mockAgentRegistrationStateRepository,
	)

	return &WhenPublishingRuntimeRecordsTestingSuite{
		sut:                                  sut,
		mockEnsRegistrar:                     mockEnsRegistrar,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockRuntimeRecordsProvider:           mockRuntimeRecordsProvider,
	}
}

func TestWhenPublishingRuntimeRecords(t *testing.T) {
	t.Parallel()

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenPublishingRuntimeRecordsTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
		}

		t.Run("Should report ErrIdentityNotFound without touching the registrar", func(t *testing.T) {
			t.Parallel()

			suite := WhenPublishingRuntimeRecordsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.PublishRuntimeRecords(context.Background())

			assert.ErrorIs(t, err, apperrors.ErrIdentityNotFound)
		})
	})

	t.Run("Given the registration state lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenPublishingRuntimeRecordsTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(5).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenPublishingRuntimeRecordsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.PublishRuntimeRecords(context.Background())

			assert.ErrorContains(t, err, "db down")
		})
	})

	t.Run("Given a minted agent whose subnode is not provisioned", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenPublishingRuntimeRecordsTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(5).WithName("alice").Build()
			state := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusConfirmed).
				Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(state, nil)
			suite.mockEnsRegistrar.EXPECT().
				Provisioned(gomock.Any(), gomock.Any(), gomock.Any(), "alice").
				Return(false, nil)
			suite.mockRuntimeRecordsProvider.EXPECT().RuntimeRecords().Return(runtimeRecords)
			suite.mockEnsRegistrar.EXPECT().
				InitRecords(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(
					func(_ context.Context, _ string, _ common.Address, _ *big.Int, records []identity_interfaces.EnsRecord) error {
						assert.Equal(t, runtimeRecords, records)
						return nil
					})
		}

		t.Run("Should seed the runtime records via InitRecords", func(t *testing.T) {
			t.Parallel()

			suite := WhenPublishingRuntimeRecordsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.PublishRuntimeRecords(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given a provisioned subnode whose url record has drifted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenPublishingRuntimeRecordsTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(5).WithName("alice").Build()
			state := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusConfirmed).
				Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(state, nil)
			suite.mockEnsRegistrar.EXPECT().
				Provisioned(gomock.Any(), gomock.Any(), gomock.Any(), "alice").
				Return(true, nil)
			suite.mockRuntimeRecordsProvider.EXPECT().RuntimeRecords().Return(runtimeRecords)
			suite.mockEnsRegistrar.EXPECT().
				Text(gomock.Any(), gomock.Any(), gomock.Any(), "alice", "url").
				Return("https://old.example.com", nil)
			suite.mockEnsRegistrar.EXPECT().
				SetRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "url", runtimeUrl).
				Return(nil)
		}

		t.Run("Should overwrite it via SetRecord with the current url", func(t *testing.T) {
			t.Parallel()

			suite := WhenPublishingRuntimeRecordsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.PublishRuntimeRecords(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given a provisioned subnode whose url record is already current", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenPublishingRuntimeRecordsTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(5).WithName("alice").Build()
			state := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusConfirmed).
				Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(state, nil)
			suite.mockEnsRegistrar.EXPECT().
				Provisioned(gomock.Any(), gomock.Any(), gomock.Any(), "alice").
				Return(true, nil)
			suite.mockRuntimeRecordsProvider.EXPECT().RuntimeRecords().Return(runtimeRecords)
			suite.mockEnsRegistrar.EXPECT().
				Text(gomock.Any(), gomock.Any(), gomock.Any(), "alice", "url").
				Return(runtimeUrl, nil)
		}

		t.Run("Should not write anything", func(t *testing.T) {
			t.Parallel()

			suite := WhenPublishingRuntimeRecordsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.PublishRuntimeRecords(context.Background())

			assert.NoError(t, err)
		})
	})
}
