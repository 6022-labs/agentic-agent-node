package use_cases_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
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
	mockGenerateAgentImages              *interface_mocks.MockGenerateAgentImagesInterface
	mockIPFSFileStorage                  *interface_mocks.MockIPFSFileStorageInterface
	mockToken6022Requester               *interface_mocks.MockToken6022RequesterInterface
}

func WhenSavingAgentRegistrationBeforeEach(t *testing.T) *WhenSavingAgentRegistrationTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(mockController)
	mockAgentCollectionsManagerRequester := interface_mocks.NewMockAgentCollectionsManagerRequesterInterface(mockController)
	mockAgentCollectionRequester := interface_mocks.NewMockAgentCollectionRequesterInterface(mockController)
	mockWalletCredentialsStorage := interface_mocks.NewMockWalletCredentialsStorageInterface(mockController)
	mockGenerateAgentImages := interface_mocks.NewMockGenerateAgentImagesInterface(mockController)
	mockIPFSFileStorage := interface_mocks.NewMockIPFSFileStorageInterface(mockController)
	mockToken6022Requester := interface_mocks.NewMockToken6022RequesterInterface(mockController)

	agentIdentitySettings := &settings.AgentIdentitySettings{
		RuntimeUrl:    "https://agent.example.com",
		EnsBaseDomain: "6022.eth",
	}
	// 84532 is reachable but has no creator, so it's a payable/rating chain but
	// not identity-capable.
	chainsSettings := &runtime_common_settings.ChainsSettings{
		"80002": {ChainId: "80002", HttpUrl: "http://localhost:8545"},
		"84532": {ChainId: "84532", HttpUrl: "http://localhost:8546"},
	}
	agentCollectionCreatorsSettings := &runtime_common_settings.AgentCollectionCreatorsSettings{
		"80002": common_tests.RandomEthAddress(),
	}

	sut := use_cases.NewSaveAgentRegistration(mockAgentRepository, mockAgentRegistrationStateRepository, mockAgentCollectionsManagerRequester, mockAgentCollectionRequester, mockWalletCredentialsStorage, mockGenerateAgentImages, mockIPFSFileStorage, mockToken6022Requester, identity_services.NewEnsDomainBuilder(agentIdentitySettings), chainsSettings, agentCollectionCreatorsSettings)

	return &WhenSavingAgentRegistrationTestingSuite{
		sut:                                  sut,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockAgentCollectionsManagerRequester: mockAgentCollectionsManagerRequester,
		mockAgentCollectionRequester:         mockAgentCollectionRequester,
		mockWalletCredentialsStorage:         mockWalletCredentialsStorage,
		mockGenerateAgentImages:              mockGenerateAgentImages,
		mockIPFSFileStorage:                  mockIPFSFileStorage,
		mockToken6022Requester:               mockToken6022Requester,
	}
}

func validSaveRequest(t *testing.T) *requests.SaveAgentRegistrationRequest {
	return &requests.SaveAgentRegistrationRequest{
		ChainId:           "80002",
		CollectionAddress: common_tests.RandomEthAddress().Hex(),
		Owner:             common_tests.RandomEthAddress().Hex(),
		Name:              "Alice",
		Role:              identity_enums.AgentRoleExpert.String(),
		Image:             newMultipartFileHeader(t, []byte("fake-image"), "image/png"),
		Attributes:        map[string]string{},
	}
}

func TestWhenSavingAgentRegistration(t *testing.T) {
	t.Parallel()

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)

			request := validSaveRequest(t)
			request.Name = ""

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the collection is known and the wallet is ready", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockAgentCollectionsManagerRequester.EXPECT().IsKnownCollection("80002", gomock.Any()).Return(true, nil)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockWalletCredentialsStorage.EXPECT().GetForId(gomock.Any(), gomock.Any()).Return(&identity_interfaces.WalletCredentials{Address: common_tests.RandomEthAddress().Hex()}, true, nil)
			suite.mockAgentCollectionRequester.EXPECT().CollectionName("80002", gomock.Any()).Return("squad", nil)
			suite.mockToken6022Requester.EXPECT().BalanceOf("80002", gomock.Any()).Return(big.NewInt(1_000_000), nil)
			// $6022 balance is summed across every reachable chain; a chain the
			// creator can't be read on is skipped (best-effort).
			suite.mockToken6022Requester.EXPECT().BalanceOf("84532", gomock.Any()).Return(nil, assert.AnError)
			suite.mockToken6022Requester.EXPECT().Decimals().Return(uint8(18), nil)
			cid := "QmCid"
			suite.mockGenerateAgentImages.EXPECT().Generate(gomock.Any()).Return(map[string][]byte{"default": []byte("a"), "icon": []byte("b")}, nil)
			suite.mockIPFSFileStorage.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return(&cid, nil).Times(2)
			suite.mockAgentRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, a *models.Agent) (*models.Agent, error) { return a, nil })
			suite.mockAgentRegistrationStateRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, s *models.AgentRegistrationState) (*models.AgentRegistrationState, error) {
					return s, nil
				})
		}

		t.Run("Should return Ok with the stored draft", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), validSaveRequest(t))

			assert.NoError(t, err)
			assert.Equal(t, "Alice", resp.Name)
			assert.Equal(t, "alice.squad.80002.6022.eth", resp.EnsDomain)
			assert.Equal(t, string(identity_enums.TransactionStatusPending), resp.MintStatus)
			assert.Equal(t, "QmCid", resp.Images["default"])
			assert.Equal(t, "QmCid", resp.Images["icon"])
		})
	})

	t.Run("Given the collection is not governed by the manager", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockAgentCollectionsManagerRequester.EXPECT().IsKnownCollection("80002", gomock.Any()).Return(false, nil)
		}

		t.Run("Should return a validation error on collectionAddress", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), validSaveRequest(t))

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "collectionAddress", validationErr.Field)
		})
	})

	t.Run("Given a chain the node has no RPC for", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on chainId without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)

			request := validSaveRequest(t)
			request.ChainId = "999999"

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "chainId", validationErr.Field)
		})
	})

	t.Run("Given a reachable chain with no collection creator", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on chainId without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)

			request := validSaveRequest(t)
			request.ChainId = "84532"

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "chainId", validationErr.Field)
		})
	})
}
