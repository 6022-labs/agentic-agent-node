package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenCheckingEnsProvisioningTestingSuite struct {
	sut *services.EnsPublisher

	mockEnsRegistrar                     *interface_mocks.MockEnsRegistrarInterface
	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
	mockRuntimeRecordsProvider           *interface_mocks.MockRuntimeRecordsProviderInterface
}

func WhenCheckingEnsProvisioningBeforeEach(t *testing.T) *WhenCheckingEnsProvisioningTestingSuite {
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

	return &WhenCheckingEnsProvisioningTestingSuite{
		sut:                                  sut,
		mockEnsRegistrar:                     mockEnsRegistrar,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockRuntimeRecordsProvider:           mockRuntimeRecordsProvider,
	}
}

func TestWhenCheckingEnsProvisioning(t *testing.T) {
	t.Parallel()

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingEnsProvisioningTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
		}

		t.Run("Should report ErrIdentityNotFound without reading the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingEnsProvisioningBeforeEach(t)
			initSuite(suite)

			provisioned, err := suite.sut.IsProvisioned(context.Background())

			assert.ErrorIs(t, err, common_interfaces.ErrIdentityNotFound)
			assert.False(t, provisioned)
		})
	})

	t.Run("Given a minted agent", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingEnsProvisioningTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(5).WithName("alice").Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockEnsRegistrar.EXPECT().
				Provisioned(gomock.Any(), gomock.Any(), gomock.Any(), "alice").
				Return(true, nil)
		}

		t.Run("Should report the registrar's provisioning result", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingEnsProvisioningBeforeEach(t)
			initSuite(suite)

			provisioned, err := suite.sut.IsProvisioned(context.Background())

			assert.NoError(t, err)
			assert.True(t, provisioned)
		})
	})
}
