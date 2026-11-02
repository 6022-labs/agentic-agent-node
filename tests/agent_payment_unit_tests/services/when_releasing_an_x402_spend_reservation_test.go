package services_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type releaseUnitOfWorkContextKey struct{}

type WhenReleasingAnX402SpendReservationTestingSuite struct {
	sut *services.X402PaymentSpendPolicyEvaluator

	mockPaymentSpendRuleWindowRepository *payment_repositories_mocks.MockPaymentSpendRuleWindowRepositoryInterface
	mockUnitOfWorkFactory                *runtime_common_mocks.MockUnitOfWorkFactoryInterface
	mockUnitOfWork                       *runtime_common_mocks.MockUnitOfWorkInterface
	unitOfWorkCtx                        context.Context
}

func WhenReleasingAnX402SpendReservationBeforeEach(t *testing.T) *WhenReleasingAnX402SpendReservationTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendRuleWindowRepository :=
		payment_repositories_mocks.NewMockPaymentSpendRuleWindowRepositoryInterface(mockController)
	mockUnitOfWorkFactory := runtime_common_mocks.NewMockUnitOfWorkFactoryInterface(mockController)
	unitOfWorkCtx := context.WithValue(context.Background(), releaseUnitOfWorkContextKey{}, "unit_of_work")
	mockUnitOfWork := runtime_common_mocks.NewMockUnitOfWorkInterface(mockController)
	mockUnitOfWork.EXPECT().Context().Return(unitOfWorkCtx).AnyTimes()
	mockUnitOfWork.EXPECT().Rollback().Return(nil).AnyTimes()

	return &WhenReleasingAnX402SpendReservationTestingSuite{
		sut: services.NewX402PaymentSpendPolicyEvaluator(
			payment_interfaces_mocks.NewMockPaymentSpendPolicyProviderInterface(mockController),
			payment_repositories_mocks.NewMockPaymentSpendRuleRepositoryInterface(mockController),
			payment_repositories_mocks.NewMockPaymentSpendRuleX402ConfigRepositoryInterface(mockController),
			mockPaymentSpendRuleWindowRepository,
			mockUnitOfWorkFactory,
		),

		mockPaymentSpendRuleWindowRepository: mockPaymentSpendRuleWindowRepository,
		mockUnitOfWorkFactory:                mockUnitOfWorkFactory,
		mockUnitOfWork:                       mockUnitOfWork,
		unitOfWorkCtx:                        unitOfWorkCtx,
	}
}

func TestWhenReleasingAnX402SpendReservation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ruleId := uuid.New()
	periodStartedAt := time.Now().Add(-time.Hour).Truncate(time.Second)

	t.Run("Given the reservation charged no window", func(t *testing.T) {
		t.Parallel()

		reservation := &payment_interfaces.X402PaymentSpendReservation{
			RuleId:       ruleId,
			Amount:       big.NewInt(10000),
			WindowStarts: map[uuid.UUID]time.Time{},
		}

		// No Start expectation: a cap-only spend has nothing to give back.
		t.Run("Should do nothing", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingAnX402SpendReservationBeforeEach(t)

			err := suite.sut.Release(ctx, reservation)

			assert.NoError(t, err)
		})
	})

	t.Run("Given starting the unit of work fails", func(t *testing.T) {
		t.Parallel()

		window := model_builders.NewPaymentSpendRuleWindowBuilder().WithPaymentSpendRuleId(ruleId).Build()
		reservation := &payment_interfaces.X402PaymentSpendReservation{
			RuleId:       ruleId,
			Amount:       big.NewInt(10000),
			WindowStarts: map[uuid.UUID]time.Time{window.Id: periodStartedAt},
		}

		initSuite := func(suite *WhenReleasingAnX402SpendReservationTestingSuite) {
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingAnX402SpendReservationBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Release(ctx, reservation)

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given reading the rule's windows fails", func(t *testing.T) {
		t.Parallel()

		window := model_builders.NewPaymentSpendRuleWindowBuilder().WithPaymentSpendRuleId(ruleId).Build()
		reservation := &payment_interfaces.X402PaymentSpendReservation{
			RuleId:       ruleId,
			Amount:       big.NewInt(10000),
			WindowStarts: map[uuid.UUID]time.Time{window.Id: periodStartedAt},
		}

		initSuite := func(suite *WhenReleasingAnX402SpendReservationTestingSuite) {
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(suite.unitOfWorkCtx, ruleId).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingAnX402SpendReservationBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Release(ctx, reservation)

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given a charged window's spend is not a usable amount", func(t *testing.T) {
		t.Parallel()

		window := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(ruleId).
			WithSpent("not-a-number").
			WithStartedAt(periodStartedAt).
			Build()
		reservation := &payment_interfaces.X402PaymentSpendReservation{
			RuleId:       ruleId,
			Amount:       big.NewInt(10000),
			WindowStarts: map[uuid.UUID]time.Time{window.Id: periodStartedAt},
		}

		initSuite := func(suite *WhenReleasingAnX402SpendReservationTestingSuite) {
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(suite.unitOfWorkCtx, ruleId).
				Return([]models.PaymentSpendRuleWindow{*window}, nil).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingAnX402SpendReservationBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Release(ctx, reservation)

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given saving a window fails", func(t *testing.T) {
		t.Parallel()

		window := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(ruleId).
			WithSpent("30000").
			WithStartedAt(periodStartedAt).
			Build()
		reservation := &payment_interfaces.X402PaymentSpendReservation{
			RuleId:       ruleId,
			Amount:       big.NewInt(10000),
			WindowStarts: map[uuid.UUID]time.Time{window.Id: periodStartedAt},
		}

		initSuite := func(suite *WhenReleasingAnX402SpendReservationTestingSuite) {
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(suite.unitOfWorkCtx, ruleId).
				Return([]models.PaymentSpendRuleWindow{*window}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingAnX402SpendReservationBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Release(ctx, reservation)

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given committing the release fails", func(t *testing.T) {
		t.Parallel()

		window := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(ruleId).
			WithSpent("30000").
			WithStartedAt(periodStartedAt).
			Build()
		reservation := &payment_interfaces.X402PaymentSpendReservation{
			RuleId:       ruleId,
			Amount:       big.NewInt(10000),
			WindowStarts: map[uuid.UUID]time.Time{window.Id: periodStartedAt},
		}

		initSuite := func(suite *WhenReleasingAnX402SpendReservationTestingSuite) {
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(suite.unitOfWorkCtx, ruleId).
				Return([]models.PaymentSpendRuleWindow{*window}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, gomock.Any()).
				Return(window, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(assert.AnError).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingAnX402SpendReservationBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Release(ctx, reservation)

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given one window rolled over since the reservation", func(t *testing.T) {
		t.Parallel()

		chargedWindow := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(ruleId).
			WithSpent("30000").
			WithStartedAt(periodStartedAt).
			Build()
		rolledWindow := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(ruleId).
			WithWindowSeconds(3600).
			WithSpent("5000").
			WithStartedAt(periodStartedAt.Add(time.Hour)).
			Build()
		reservation := &payment_interfaces.X402PaymentSpendReservation{
			RuleId: ruleId,
			Amount: big.NewInt(10000),
			WindowStarts: map[uuid.UUID]time.Time{
				chargedWindow.Id: periodStartedAt,
				rolledWindow.Id:  periodStartedAt,
			},
		}

		initSuite := func(suite *WhenReleasingAnX402SpendReservationTestingSuite) {
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(suite.unitOfWorkCtx, ruleId).
				Return([]models.PaymentSpendRuleWindow{*chargedWindow, *rolledWindow}, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)
		}

		// The new period never held the amount, so taking it back there would free budget twice.
		t.Run("Should leave the new period's spend unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingAnX402SpendReservationBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, windowWithSpend{id: chargedWindow.Id.String(), spent: "20000"}).
				Return(chargedWindow, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, windowWithSpend{id: rolledWindow.Id.String(), spent: "5000"}).
				Return(rolledWindow, nil).
				Times(1)

			err := suite.sut.Release(ctx, reservation)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the reservation's windows are still in their period", func(t *testing.T) {
		t.Parallel()

		dailyWindow := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(ruleId).
			WithSpent("30000").
			WithStartedAt(periodStartedAt).
			Build()
		weeklyWindow := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(ruleId).
			WithWindowSeconds(604800).
			WithSpent("10000").
			WithStartedAt(periodStartedAt).
			Build()
		uncharged := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(ruleId).
			WithWindowSeconds(3600).
			WithSpent("7000").
			WithStartedAt(periodStartedAt).
			Build()
		reservation := &payment_interfaces.X402PaymentSpendReservation{
			RuleId: ruleId,
			Amount: big.NewInt(10000),
			WindowStarts: map[uuid.UUID]time.Time{
				dailyWindow.Id:  periodStartedAt,
				weeklyWindow.Id: periodStartedAt,
			},
		}

		initSuite := func(suite *WhenReleasingAnX402SpendReservationTestingSuite) {
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(suite.unitOfWorkCtx, ruleId).
				Return([]models.PaymentSpendRuleWindow{*dailyWindow, *weeklyWindow, *uncharged}, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)
		}

		// No Update expectation for the uncharged window: a window the reservation never touched stays as is.
		t.Run("Should give the amount back to every charged window", func(t *testing.T) {
			t.Parallel()

			suite := WhenReleasingAnX402SpendReservationBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, windowWithSpend{id: dailyWindow.Id.String(), spent: "20000"}).
				Return(dailyWindow, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, windowWithSpend{id: weeklyWindow.Id.String(), spent: "0"}).
				Return(weeklyWindow, nil).
				Times(1)

			err := suite.sut.Release(ctx, reservation)

			assert.NoError(t, err)
		})
	})
}
