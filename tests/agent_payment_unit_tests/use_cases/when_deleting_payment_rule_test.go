package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingPaymentRuleTestingSuite struct {
	sut *use_cases.DeletePaymentRule

	mockPaymentRuleRepository *payment_repositories_mocks.MockPaymentRuleRepositoryInterface
}

func WhenDeletingPaymentRuleBeforeEach(t *testing.T) *WhenDeletingPaymentRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentRuleRepository := payment_repositories_mocks.NewMockPaymentRuleRepositoryInterface(mockController)

	return &WhenDeletingPaymentRuleTestingSuite{
		sut: use_cases.NewDeletePaymentRule(mockPaymentRuleRepository),

		mockPaymentRuleRepository: mockPaymentRuleRepository,
	}
}

func TestWhenDeletingPaymentRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the rule exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentRuleTestingSuite) *uuid.UUID {
			rule := model_builders.NewPaymentRuleBuilder().Build()

			suite.mockPaymentRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().DeleteById(gomock.Any(), rule.Id).Return(nil).Times(1)
			return &rule.Id
		}

		t.Run("Should delete the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentRuleBeforeEach(t)
			id := initSuite(suite)

			assert.NoError(t, suite.sut.Execute(context.Background(), &requests.DeletePaymentRuleRequest{Id: *id}))
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleRepository.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		}

		t.Run("Should return a not found error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentRuleRequest{Id: uuid.New()})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})
}
