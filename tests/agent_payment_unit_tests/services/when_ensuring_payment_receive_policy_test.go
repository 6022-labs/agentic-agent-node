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
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
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

	mockAgentRepository                *identity_repositories_mocks.MockAgentRepositoryInterface
	mockAgentAddressRepository         *identity_repositories_mocks.MockAgentAddressRepositoryInterface
	mockPaymentReceivePolicyRepository *payment_repositories_mocks.MockPaymentReceivePolicyRepositoryInterface
	mockPaymentReceiveRuleRepository   *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
	mockStrategy                       *payment_interfaces_mocks.MockPaymentReceiveMethodStrategyInterface
}

func WhenEnsuringPaymentReceivePolicyBeforeEach(t *testing.T) *WhenEnsuringPaymentReceivePolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentAddressRepository := identity_repositories_mocks.NewMockAgentAddressRepositoryInterface(mockController)
	mockPaymentReceivePolicyRepository := payment_repositories_mocks.NewMockPaymentReceivePolicyRepositoryInterface(
		mockController,
	)
	mockPaymentReceiveRuleRepository := payment_repositories_mocks.NewMockPaymentReceiveRuleRepositoryInterface(
		mockController,
	)
	mockStrategy := payment_interfaces_mocks.NewMockPaymentReceiveMethodStrategyInterface(mockController)
	mockStrategies := payment_interfaces_mocks.NewMockPaymentReceiveMethodStrategiesInterface(mockController)
	mockStrategies.EXPECT().For(gomock.Any()).Return(mockStrategy, true).AnyTimes()

	token6022Settings := &runtime_common_settings.Token6022Settings{
		receiveSeedChainId: common.HexToAddress(receiveSeedAsset),
	}

	return &WhenEnsuringPaymentReceivePolicyTestingSuite{
		sut: services.NewPaymentReceivePolicyProvider(
			mockAgentRepository,
			mockAgentAddressRepository,
			mockPaymentReceivePolicyRepository,
			mockPaymentReceiveRuleRepository,
			mockStrategies,
			token6022Settings,
		),

		mockAgentRepository:                mockAgentRepository,
		mockAgentAddressRepository:         mockAgentAddressRepository,
		mockPaymentReceivePolicyRepository: mockPaymentReceivePolicyRepository,
		mockPaymentReceiveRuleRepository:   mockPaymentReceiveRuleRepository,
		mockStrategy:                       mockStrategy,
	}
}

// expectWallet stubs the agent's evm address lookup the seed scopes its rule to.
func (s *WhenEnsuringPaymentReceivePolicyTestingSuite) expectWallet(address string) {
	s.mockAgentAddressRepository.EXPECT().
		GetByAgentIdAndType(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm).
		Return(&identity_models.AgentAddress{Address: address, Type: identity_enums.AddressTypeEvm}, nil).
		AnyTimes()
}

func TestWhenEnsuringPaymentReceivePolicy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the agent row does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(nil, nil).Times(1)
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
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
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

	t.Run("Given a concurrent request won the insert race", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) *models.PaymentReceivePolicy {
			agent := model_builders.NewAgentBuilder().Build()
			winner := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				Create(ctx, gomock.Any()).
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

	t.Run("Given seeding the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithChainId(receiveSeedChainId).Build()
			created := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().Create(ctx, gomock.Any()).Return(created, nil).Times(1)
			suite.expectWallet(receiveSeedWallet)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			policy, _, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given the agent already has a policy", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) *models.PaymentReceivePolicy {
			agent := model_builders.NewAgentBuilder().Build()
			existing := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
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

	t.Run("Given the agent has neither a wallet nor an owner", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentReceivePolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithChainId(receiveSeedChainId).Build()
			agent.Owner = nil
			created := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().Create(ctx, gomock.Any()).Return(created, nil).Times(1)
			suite.mockAgentAddressRepository.EXPECT().
				GetByAgentIdAndType(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm).
				Return(nil, nil).
				Times(1)
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

	t.Run("Given the agent has no policy yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(
			suite *WhenEnsuringPaymentReceivePolicyTestingSuite,
		) (**models.PaymentReceiveRule, *json.RawMessage) {
			agent := model_builders.NewAgentBuilder().
				WithChainId(receiveSeedChainId).
				WithOwner(receiveSeedOwner).
				Build()
			created := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			seededRule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(created.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().Create(ctx, gomock.Any()).Return(created, nil).Times(1)
			suite.expectWallet(receiveSeedWallet)

			seeded := new(*models.PaymentReceiveRule)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentReceiveRule) (*models.PaymentReceiveRule, error) {
					*seeded = rule
					return seededRule, nil
				}).
				Times(1)

			seededConfig := new(json.RawMessage)
			suite.mockStrategy.EXPECT().
				SaveConfig(ctx, seededRule.Id, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ uuid.UUID, raw json.RawMessage) error {
					*seededConfig = raw
					return nil
				}).
				Times(1)
			return seeded, seededConfig
		}

		t.Run("Should seed a free rule so the owner never pays their own agent", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			seeded, _ := initSuite(suite)

			_, _, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			assert.Equal(t, "0", (*seeded).Price)
			assert.Equal(t, "Owner free access", (*seeded).Name)
			assert.Equal(t, enums.PaymentMethodX402, (*seeded).Method)
		})

		t.Run("Should scope the seed to the agent wallet and the owner, lowercased", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentReceivePolicyBeforeEach(t)
			_, seededConfig := initSuite(suite)

			_, _, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			config := &models.PaymentReceiveRuleX402Config{}
			assert.NoError(t, config.Parse(*seededConfig))
			// Patterns are matched against the lowercased address, so the seed must store them that way.
			assert.Equal(t, []string{
				strings.ToLower(receiveSeedWallet),
				strings.ToLower(receiveSeedOwner),
			}, config.AddressPatterns)
			assert.Equal(t, "eip155:80002", config.Network)
			assert.Equal(t, receiveSeedAsset, config.Asset)
			assert.Equal(t, receiveSeedWallet, config.PayTo)
			assert.Equal(t, enums.AssetTransferMethodEip3009, config.AssetTransferMethod)
		})
	})
}
