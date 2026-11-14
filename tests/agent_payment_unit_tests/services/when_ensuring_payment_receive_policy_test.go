package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	identity_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const receiveSeedChainId = "80002"

var (
	receiveSeedWallet = common_tests.RandomEthAddress().Hex()
	receiveSeedOwner  = common_tests.RandomEthAddress().Hex()
	receiveSeedAsset  = common_tests.RandomEthAddress().Hex()
)

type WhenEnsuringPaymentReceivePolicyTestingSuite struct {
	sut *services.PaymentReceivePolicyProvider

	mockAgentResolver                     *identity_interfaces_mocks.MockAgentResolverInterface
	mockAgentAddressRepository            *identity_repositories_mocks.MockAgentAddressRepositoryInterface
	mockPaymentReceivePolicyRepository    *payment_repositories_mocks.MockPaymentReceivePolicyRepositoryInterface
	mockPaymentReceiveRuleRepository      *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
	mockPaymentReceiveRulePriceRepository *payment_repositories_mocks.MockPaymentReceiveRulePriceRepositoryInterface
	mockStrategy                          *payment_interfaces_mocks.MockPaymentReceiveMethodStrategyInterface
	mockUnitOfWorkFactory                 *runtime_common_mocks.MockUnitOfWorkFactoryInterface
	mockUnitOfWork                        *runtime_common_mocks.MockUnitOfWorkInterface
	unitOfWorkCtx                         context.Context
}

type receiveSeedUnitOfWorkContextKey struct{}

func WhenEnsuringPaymentReceivePolicyBeforeEach(t *testing.T) *WhenEnsuringPaymentReceivePolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentResolver := identity_interfaces_mocks.NewMockAgentResolverInterface(mockController)
	mockAgentAddressRepository := identity_repositories_mocks.NewMockAgentAddressRepositoryInterface(mockController)
	mockPaymentReceivePolicyRepository := payment_repositories_mocks.NewMockPaymentReceivePolicyRepositoryInterface(
		mockController,
	)
	mockPaymentReceiveRuleRepository := payment_repositories_mocks.NewMockPaymentReceiveRuleRepositoryInterface(
		mockController,
	)
	mockPaymentReceiveRulePriceRepository := payment_repositories_mocks.NewMockPaymentReceiveRulePriceRepositoryInterface(
		mockController,
	)
	mockStrategy := payment_interfaces_mocks.NewMockPaymentReceiveMethodStrategyInterface(mockController)
	mockStrategies := payment_interfaces_mocks.NewMockPaymentReceiveMethodStrategiesInterface(mockController)
	mockStrategies.EXPECT().For(gomock.Any()).Return(mockStrategy, true).AnyTimes()

	mockUnitOfWorkFactory := runtime_common_mocks.NewMockUnitOfWorkFactoryInterface(mockController)
	unitOfWorkCtx := context.WithValue(context.Background(), receiveSeedUnitOfWorkContextKey{}, "unit_of_work")
	mockUnitOfWork := runtime_common_mocks.NewMockUnitOfWorkInterface(mockController)
	mockUnitOfWork.EXPECT().Context().Return(unitOfWorkCtx).AnyTimes()
	mockUnitOfWork.EXPECT().Rollback().Return(nil).AnyTimes()

	token6022Settings := &runtime_common_settings.Token6022Settings{
		receiveSeedChainId: common.HexToAddress(receiveSeedAsset),
	}

	return &WhenEnsuringPaymentReceivePolicyTestingSuite{
		sut: services.NewPaymentReceivePolicyProvider(
			mockAgentResolver,
			mockAgentAddressRepository,
			mockPaymentReceivePolicyRepository,
			mockPaymentReceiveRuleRepository,
			mockPaymentReceiveRulePriceRepository,
			mockStrategies,
			token6022Settings,
			mockUnitOfWorkFactory,
		),

		mockAgentResolver:                     mockAgentResolver,
		mockAgentAddressRepository:            mockAgentAddressRepository,
		mockPaymentReceivePolicyRepository:    mockPaymentReceivePolicyRepository,
		mockPaymentReceiveRuleRepository:      mockPaymentReceiveRuleRepository,
		mockPaymentReceiveRulePriceRepository: mockPaymentReceiveRulePriceRepository,
		mockStrategy:                          mockStrategy,
		mockUnitOfWorkFactory:                 mockUnitOfWorkFactory,
		mockUnitOfWork:                        mockUnitOfWork,
		unitOfWorkCtx:                         unitOfWorkCtx,
	}
}

// expectWallet stubs the agent's evm address lookup the seed scopes its rule to.
func (s *WhenEnsuringPaymentReceivePolicyTestingSuite) expectWallet(address string) {
	s.mockAgentAddressRepository.EXPECT().
		GetByAgentIdAndType(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm).
		Return(&identity_models.AgentAddress{Address: address, Type: identity_enums.AddressTypeEvm}, nil).
		AnyTimes()
}

// expectNoPolicy stubs an agent with no policy yet, read again after a failed creation.
func (s *WhenEnsuringPaymentReceivePolicyTestingSuite) expectNoPolicy(agent *identity_models.Agent, reads int) {
	s.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
	s.mockPaymentReceivePolicyRepository.EXPECT().GetByAgentId(gomock.Any(), agent.Id).Return(nil, nil).Times(reads)
}

func TestWhenEnsuringPaymentReceivePolicy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the agent row does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			suite.mockAgentResolver.EXPECT().Get(ctx).Return(nil, nil).Times(1)
		}

		t.Run("Should fail rather than create an orphan policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			policy, _, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			suite.mockAgentResolver.EXPECT().Get(ctx).Return(agent, nil).Times(1)
		}

		// The seed prices access to the agent's token, so there is nothing to seed before mint.
		t.Run("Should return a conflict without touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			policy, _, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var conflictError *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictError)
			assert.Equal(t, "agent_not_minted", conflictError.Code)
		})
	})

	t.Run("Given the agent already has a policy", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) *models.PaymentReceivePolicy {
			agent := model_builders.NewAgentBuilder().Build()
			existing := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentResolver.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				GetByAgentId(ctx, agent.Id).
				Return(existing, nil).
				Times(1)
			return existing
		}

		t.Run("Should return it without seeding a second rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			existing := initSuite(suite)

			policy, agent, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			assert.NotNil(t, agent)
			assert.Equal(t, existing.Id, policy.Id)
		})
	})

	t.Run("Given reading the agent wallet fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			suite.expectNoPolicy(model_builders.NewAgentBuilder().Build(), 2)
			suite.mockAgentAddressRepository.EXPECT().
				GetByAgentIdAndType(ctx, gomock.Any(), identity_enums.AddressTypeEvm).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError without opening a unit of work", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			policy, _, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "policy_failed", internalError.Code)
		})
	})

	t.Run("Given the unit of work cannot start", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			suite.expectNoPolicy(model_builders.NewAgentBuilder().Build(), 2)
			suite.expectWallet(receiveSeedWallet)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return an InternalError without creating the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			policy, _, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "policy_failed", internalError.Code)
		})
	})

	t.Run("Given a concurrent request won the insert race", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) *models.PaymentReceivePolicy {
			agent := model_builders.NewAgentBuilder().Build()
			winner := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentResolver.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.expectWallet(receiveSeedWallet)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(nil, errors.New("duplicate key")).
				Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(winner, nil).Times(1)
			return winner
		}

		t.Run("Should return the row that won rather than fail", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			winner := initSuite(suite)

			policy, _, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			assert.Equal(t, winner.Id, policy.Id)
		})
	})

	t.Run("Given the agent has neither a wallet nor an owner", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithChainId(receiveSeedChainId).Build()
			agent.Owner = nil
			created := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.expectNoPolicy(agent, 1)
			suite.mockAgentAddressRepository.EXPECT().
				GetByAgentIdAndType(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm).
				Return(nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(created, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)
		}

		// A rule scoped to nobody would be a rule anyone matches, so no rule is seeded at all.
		t.Run("Should create the policy without seeding a rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			policy, _, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			assert.NotNil(t, policy)
		})
	})

	t.Run("Given seeding the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithChainId(receiveSeedChainId).Build()
			created := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.expectNoPolicy(agent, 2)
			suite.expectWallet(receiveSeedWallet)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(created, nil).
				Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Times(0)
		}

		// Nothing is committed, so the next call finds no policy and seeds again.
		t.Run("Should return an InternalError without committing the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			policy, _, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "policy_failed", internalError.Code)
		})
	})

	t.Run("Given seeding the price fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithChainId(receiveSeedChainId).Build()
			created := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			seededRule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(created.Id).Build()

			suite.expectNoPolicy(agent, 2)
			suite.expectWallet(receiveSeedWallet)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(created, nil).
				Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(seededRule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(suite.unitOfWorkCtx, created.Id, seededRule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Times(0)
		}

		t.Run("Should return an InternalError without committing the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			policy, _, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "policy_failed", internalError.Code)
		})
	})

	t.Run("Given committing the seed fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithChainId(receiveSeedChainId).Build()
			created := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			seededRule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(created.Id).Build()

			suite.expectNoPolicy(agent, 2)
			suite.expectWallet(receiveSeedWallet)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(created, nil).
				Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(seededRule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(suite.unitOfWorkCtx, created.Id, seededRule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					price *models.PaymentReceiveRulePrice,
				) (*models.PaymentReceiveRulePrice, error) {
					return price, nil
				}).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(assert.AnError).Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			policy, _, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "policy_failed", internalError.Code)
		})
	})

	t.Run("Given the agent has no policy yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(
			suite *WhenEnsuringPaymentReceivePolicyTestingSuite,
			ruleMatcher gomock.Matcher,
			configMatcher gomock.Matcher,
			priceMatcher gomock.Matcher,
		) {
			agent := model_builders.NewAgentBuilder().
				WithChainId(receiveSeedChainId).
				WithOwner(receiveSeedOwner).
				Build()
			created := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			seededRule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(created.Id).Build()

			suite.expectNoPolicy(agent, 1)
			suite.expectWallet(receiveSeedWallet)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(created, nil).
				Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(suite.unitOfWorkCtx, ruleMatcher).
				Return(seededRule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(suite.unitOfWorkCtx, created.Id, seededRule.Id, configMatcher).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(suite.unitOfWorkCtx, priceMatcher).
				DoAndReturn(func(
					_ context.Context,
					price *models.PaymentReceiveRulePrice,
				) (*models.PaymentReceiveRulePrice, error) {
					return price, nil
				}).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)
		}

		t.Run("Should seed a free rule so the owner never pays their own agent", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(
				suite,
				gomock.Cond(func(rule *models.PaymentReceiveRule) bool {
					return rule.Name == "Owner free access" && rule.Method == enums.PaymentMethodX402
				}),
				gomock.Any(),
				gomock.Cond(func(price *models.PaymentReceiveRulePrice) bool { return price.Price == "0" }),
			)

			_, _, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
		})

		t.Run("Should scope the seed to the agent wallet and the owner, lowercased", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			// Patterns are stored the way a lowercased payer address is written.
			initSuite(
				suite,
				gomock.Any(),
				gomock.Any(),
				gomock.Cond(func(price *models.PaymentReceiveRulePrice) bool {
					return assert.ObjectsAreEqual([]string{
						strings.ToLower(receiveSeedWallet),
						strings.ToLower(receiveSeedOwner),
					}, price.AddressPatterns)
				}),
			)

			_, _, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
		})

		t.Run("Should price the agent's token and pay the agent wallet", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(
				suite,
				gomock.Any(),
				gomock.Cond(func(raw json.RawMessage) bool {
					config := &models.PaymentReceiveRuleX402Config{}
					return config.Parse(raw) == nil &&
						config.Network == "eip155:80002" &&
						config.Asset == strings.ToLower(receiveSeedAsset) &&
						config.PayTo == receiveSeedWallet &&
						config.AssetTransferMethod == enums.AssetTransferMethodEip3009
				}),
				gomock.Any(),
			)

			_, _, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
		})
	})
}
