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

type WhenDeletingPaymentSpendRuleTestingSuite struct {
	sut *use_cases.DeletePaymentSpendRule

	mockPaymentSpendRuleRepository *payment_repositories_mocks.MockPaymentSpendRuleRepositoryInterface
}

func WhenDeletingPaymentSpendRuleBeforeEach(t *testing.T) *WhenDeletingPaymentSpendRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendRuleRepository := payment_repositories_mocks.NewMockPaymentSpendRuleRepositoryInterface(
		mockController,
	)

	return &WhenDeletingPaymentSpendRuleTestingSuite{
		sut: use_cases.NewDeletePaymentSpendRule(mockPaymentSpendRuleRepository),

		mockPaymentSpendRuleRepository: mockPaymentSpendRuleRepository,
	}
}

func TestWhenDeletingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the rule lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) {
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentSpendRuleRequest{Id: uuid.New()})

			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "delete_failed", internalErr.Code)
			}
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) {
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should return a not found error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentSpendRuleRequest{Id: uuid.New()})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "rule_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given deleting the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) *uuid.UUID {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				DeleteById(gomock.Any(), rule.Id).
				Return(assert.AnError).
				Times(1)
			return &rule.Id
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			id := initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentSpendRuleRequest{Id: *id})

			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "delete_failed", internalErr.Code)
			}
		})
	})

	t.Run("Given the rule exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) *uuid.UUID {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().DeleteById(gomock.Any(), rule.Id).Return(nil).Times(1)
			return &rule.Id
		}

		t.Run("Should delete the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			id := initSuite(suite)

			assert.NoError(
				t,
				suite.sut.Execute(context.Background(), &requests.DeletePaymentSpendRuleRequest{Id: *id}),
			)
		})
	})
}
