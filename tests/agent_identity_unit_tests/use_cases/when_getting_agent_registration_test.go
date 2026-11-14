package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	identity_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentRegistrationTestingSuite struct {
	sut *use_cases.GetAgentRegistration

	mockAgentResolver                    *identity_interfaces_mocks.MockAgentResolverInterface
	mockAgentAddressRepository           *repository_mocks.MockAgentAddressRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
}

func WhenGettingAgentRegistrationBeforeEach(t *testing.T) *WhenGettingAgentRegistrationTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentResolver := identity_interfaces_mocks.NewMockAgentResolverInterface(mockController)
	mockAgentAddressRepository := repository_mocks.NewMockAgentAddressRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(
		mockController,
	)

	sut := use_cases.NewGetAgentRegistration(
		mockAgentResolver,
		mockAgentAddressRepository,
		mockAgentRegistrationStateRepository,
	)

	return &WhenGettingAgentRegistrationTestingSuite{
		sut:                                  sut,
		mockAgentResolver:                    mockAgentResolver,
		mockAgentAddressRepository:           mockAgentAddressRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
	}
}

func TestWhenGettingAgentRegistration(t *testing.T) {
	t.Parallel()

	t.Run("Given no registration draft exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentRegistrationTestingSuite) {
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return a NotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given an agent row but no registration state", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentRegistrationTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, nil)
		}

		t.Run("Should return a NotFound since no draft has been started", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given a registration draft exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentRegistrationTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithName("alice").Build()
			state := model_builders.NewAgentRegistrationStateBuilder().Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(state, nil)
			suite.mockAgentAddressRepository.EXPECT().
				GetAllByAgentId(gomock.Any(), gomock.Any()).
				Return(map[enums.AddressType]string{enums.AddressTypeEvm: "0xabc"}, nil)
		}

		t.Run("Should return the stored draft and status", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	})
}
