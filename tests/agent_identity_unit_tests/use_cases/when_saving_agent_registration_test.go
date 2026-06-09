package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSavingAgentRegistrationTestingSuite struct {
	sut *use_cases.SaveAgentRegistration

	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
	mockAgentCollectionsManagerRequester *interface_mocks.MockAgentCollectionsManagerRequesterInterface
	mockAgentCollectionRequester         *interface_mocks.MockAgentCollectionRequesterInterface
	mockWalletCredentialsStorage         *interface_mocks.MockWalletCredentialsStorageInterface
}

func WhenSavingAgentRegistrationBeforeEach(t *testing.T) *WhenSavingAgentRegistrationTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(mockController)
	mockAgentCollectionsManagerRequester := interface_mocks.NewMockAgentCollectionsManagerRequesterInterface(mockController)
	mockAgentCollectionRequester := interface_mocks.NewMockAgentCollectionRequesterInterface(mockController)
	mockWalletCredentialsStorage := interface_mocks.NewMockWalletCredentialsStorageInterface(mockController)

	agentIdentitySettings := &settings.AgentIdentitySettings{
		RuntimeUrl:    "https://agent.example.com",
		EnsBaseDomain: "6022.eth",
	}

	sut := use_cases.NewSaveAgentRegistration(mockAgentRepository, mockAgentRegistrationStateRepository, mockAgentCollectionsManagerRequester, mockAgentCollectionRequester, mockWalletCredentialsStorage, agentIdentitySettings)

	return &WhenSavingAgentRegistrationTestingSuite{
		sut:                                  sut,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockAgentCollectionsManagerRequester: mockAgentCollectionsManagerRequester,
		mockAgentCollectionRequester:         mockAgentCollectionRequester,
		mockWalletCredentialsStorage:         mockWalletCredentialsStorage,
	}
}

func validSaveRequest() *requests.SaveAgentRegistrationRequest {
	return &requests.SaveAgentRegistrationRequest{
		CollectionAddress: common_tests.RandomEthAddress().Hex(),
		Owner:             common_tests.RandomEthAddress().Hex(),
		Name:              "Alice",
		Role:              identity_enums.AgentRoleExpert.String(),
		Images:            map[string]string{"default": "https://img"},
		Attributes:        map[string]string{},
	}
}

func TestWhenSavingAgentRegistration(t *testing.T) {
	t.Parallel()

	t.Run("Given the collection is known and the wallet is ready", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockAgentCollectionsManagerRequester.EXPECT().IsKnownCollection(gomock.Any()).Return(true, nil)
			suite.mockWalletCredentialsStorage.EXPECT().Get(gomock.Any()).Return(&identity_interfaces.WalletCredentials{Address: common_tests.RandomEthAddress().Hex()}, true, nil)
			suite.mockAgentCollectionRequester.EXPECT().CollectionName(gomock.Any()).Return("squad", nil)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
			suite.mockAgentRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, a *models.Agent) (*models.Agent, error) { return a, nil })
			suite.mockAgentRegistrationStateRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, s *models.AgentRegistrationState) (*models.AgentRegistrationState, error) {
					return s, nil
				})
		}

		t.Run("Should return Ok with the stored draft", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), validSaveRequest())

			assert.NoError(t, err)
			assert.Equal(t, "Alice", resp.Name)
			assert.Equal(t, string(identity_enums.TransactionStatusPending), resp.MintStatus)
		})
	})

	t.Run("Given the collection is not governed by the manager", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockAgentCollectionsManagerRequester.EXPECT().IsKnownCollection(gomock.Any()).Return(false, nil)
		}

		t.Run("Should return a validation error on collectionAddress", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), validSaveRequest())

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "collectionAddress", validationErr.Field)
		})
	})
}
