package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingPaymentRuleTestingSuite struct {
	sut *use_cases.UpdatePaymentRule

	mockAgentRepository         *identity_repositories_mocks.MockAgentRepositoryInterface
	mockAgentAddressRepository  *identity_repositories_mocks.MockAgentAddressRepositoryInterface
	mockPaymentPolicyRepository *payment_repositories_mocks.MockPaymentPolicyRepositoryInterface
	mockPaymentRuleRepository   *payment_repositories_mocks.MockPaymentRuleRepositoryInterface
	mockStrategy                *payment_interfaces_mocks.MockPaymentMethodStrategyInterface
}

func WhenUpdatingPaymentRuleBeforeEach(t *testing.T) *WhenUpdatingPaymentRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentAddressRepository := identity_repositories_mocks.NewMockAgentAddressRepositoryInterface(mockController)
	mockPaymentPolicyRepository := payment_repositories_mocks.NewMockPaymentPolicyRepositoryInterface(mockController)
	mockPaymentRuleRepository := payment_repositories_mocks.NewMockPaymentRuleRepositoryInterface(mockController)

	token6022Settings := &runtime_common_settings.Token6022Settings{
		"80002": common.HexToAddress("0x4242424242424242424242424242424242424242"),
	}
	strategies, mockStrategy := newMockStrategies(mockController)

	return &WhenUpdatingPaymentRuleTestingSuite{
		sut: use_cases.NewUpdatePaymentRule(
			mockAgentRepository,
			mockAgentAddressRepository,
			mockPaymentPolicyRepository,
			mockPaymentRuleRepository,
			strategies,
			token6022Settings,
		),

		mockAgentRepository:         mockAgentRepository,
		mockAgentAddressRepository:  mockAgentAddressRepository,
		mockPaymentPolicyRepository: mockPaymentPolicyRepository,
		mockPaymentRuleRepository:   mockPaymentRuleRepository,
		mockStrategy:                mockStrategy,
	}
}

func TestWhenUpdatingPaymentRule(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)

			request := generateSavePaymentRuleRequest()
			request.Name = ""
			_, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdatePaymentRuleRequest{Id: uuid.New(), SavePaymentRuleRequest: *request},
			)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given an unsupported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the method field", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)
			sut := use_cases.NewUpdatePaymentRule(
				suite.mockAgentRepository,
				suite.mockAgentAddressRepository,
				suite.mockPaymentPolicyRepository,
				suite.mockPaymentRuleRepository,
				newUnsupportedMethodStrategies(gomock.NewController(t)),
				&runtime_common_settings.Token6022Settings{},
			)

			rule, err := sut.Execute(
				context.Background(),
				&requests.UpdatePaymentRuleRequest{
					Id:                     uuid.New(),
					SavePaymentRuleRequest: *generateSavePaymentRuleRequest(),
				},
			)

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "method", validationErr.Field)
		})
	})

	t.Run("Given the strategy rejects the config", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentRuleTestingSuite) {
			suite.mockStrategy.EXPECT().
				Validate(gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("config.network", "unsupported")).
				Times(1)
		}

		t.Run("Should return a validation error before any lookup", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			updated, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdatePaymentRuleRequest{
					Id:                     uuid.New(),
					SavePaymentRuleRequest: *generateSavePaymentRuleRequest(),
				},
			)

			assert.Nil(t, updated)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.network", validationErr.Field)
		})
	})

	t.Run("Given the rule lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().
				GetById(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			updated, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdatePaymentRuleRequest{
					Id:                     uuid.New(),
					SavePaymentRuleRequest: *generateSavePaymentRuleRequest(),
				},
			)

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		}

		t.Run("Should return a not found error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			updated, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdatePaymentRuleRequest{
					Id:                     uuid.New(),
					SavePaymentRuleRequest: *generateSavePaymentRuleRequest(),
				},
			)

			assert.Nil(t, updated)
			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given saving the update fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentRuleTestingSuite) *models.PaymentRule {
			rule := model_builders.NewPaymentRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdatePaymentRuleRequest{
					Id:                     rule.Id,
					SavePaymentRuleRequest: *generateSavePaymentRuleRequest(),
				},
			)

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given the rule exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentRuleTestingSuite) *models.PaymentRule {
			rule := model_builders.NewPaymentRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, updated *models.PaymentRule) (*models.PaymentRule, error) {
					// Identity and policy binding are preserved.
					assert.Equal(t, rule.Id, updated.Id)
					assert.Equal(t, rule.PolicyId, updated.PolicyId)
					assert.Equal(t, "Friends discount", updated.Name)
					return updated, nil
				}).
				Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), rule.Id, gomock.Any()).Return(nil).Times(1)
			return rule
		}

		t.Run("Should replace the rule's fields and its config", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)
			rule := initSuite(suite)

			request := generateSavePaymentRuleRequest()
			request.Name = "Friends discount"
			updated, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdatePaymentRuleRequest{Id: rule.Id, SavePaymentRuleRequest: *request},
			)

			assert.NoError(t, err)
			assert.Equal(t, "Friends discount", updated.Name)
		})
	})
}
