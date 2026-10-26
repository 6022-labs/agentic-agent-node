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

type WhenDeletingPaymentReceiveRuleTestingSuite struct {
	sut *use_cases.DeletePaymentReceiveRule

	mockPaymentReceiveRuleRepository *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
}

func WhenDeletingPaymentReceiveRuleBeforeEach(t *testing.T) *WhenDeletingPaymentReceiveRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceiveRuleRepository := payment_repositories_mocks.NewMockPaymentReceiveRuleRepositoryInterface(
		mockController,
	)

	return &WhenDeletingPaymentReceiveRuleTestingSuite{
		sut: use_cases.NewDeletePaymentReceiveRule(mockPaymentReceiveRuleRepository),

		mockPaymentReceiveRuleRepository: mockPaymentReceiveRuleRepository,
	}
}

func TestWhenDeletingPaymentReceiveRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should return a not found error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentReceiveRuleRequest{Id: uuid.New()})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given the rule exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceiveRuleTestingSuite) *uuid.UUID {
			rule := model_builders.NewPaymentReceiveRuleBuilder().Build()

			suite.mockPaymentReceiveRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().DeleteById(gomock.Any(), rule.Id).Return(nil).Times(1)
			return &rule.Id
		}

		t.Run("Should delete the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceiveRuleBeforeEach(t)
			id := initSuite(suite)

			assert.NoError(
				t,
				suite.sut.Execute(context.Background(), &requests.DeletePaymentReceiveRuleRequest{Id: *id}),
			)
		})
	})
}
