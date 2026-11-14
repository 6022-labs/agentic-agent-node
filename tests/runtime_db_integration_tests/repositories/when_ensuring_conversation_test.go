package repositories_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type WhenEnsuringConversationTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenEnsuringConversationBeforeEach(tc *common_tests.DatabaseTestCase) *WhenEnsuringConversationTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	return &WhenEnsuringConversationTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenEnsuringConversationAfterEach(suite *WhenEnsuringConversationTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenEnsuringConversation(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the same conversation is ensured twice", func(t *testing.T) {
				t.Parallel()

				t.Run("Should keep a single row", func(t *testing.T) {
					t.Parallel()
					suite := WhenEnsuringConversationBeforeEach(tc)
					defer WhenEnsuringConversationAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					conversation := model_builders.NewConversationBuilder().
						WithId("conv-dup").
						WithAgentId(agentId).
						Build()

					created, err := suite.sut.Ensure(context.Background(), conversation)
					assert.NoError(t, err)
					assert.True(t, created)

					created, err = suite.sut.Ensure(context.Background(), conversation)
					assert.NoError(t, err)
					assert.False(t, created)

					list, err := suite.sut.ListByAgentId(context.Background(), agentId, nil, nil, nil)
					assert.NoError(t, err)
					assert.Len(t, list, 1)
				})
			})

			t.Run("Given a caller id of exactly the maximum length", func(t *testing.T) {
				t.Parallel()

				t.Run("Should store it intact", func(t *testing.T) {
					t.Parallel()
					suite := WhenEnsuringConversationBeforeEach(tc)
					defer WhenEnsuringConversationAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					callerId := strings.Repeat("é", helpers.CallerIdMaxLength)

					created, err := suite.sut.Ensure(context.Background(), model_builders.NewConversationBuilder().
						WithId("conv-wide").
						WithAgentId(agentId).
						WithCallerId(callerId).
						Build())
					assert.NoError(t, err)
					assert.True(t, created)

					stored, err := suite.sut.GetById(context.Background(), "conv-wide")
					assert.NoError(t, err)
					if assert.NotNil(t, stored) && assert.NotNil(t, stored.CallerId) {
						assert.Equal(t, callerId, *stored.CallerId)
					}
				})
			})

			t.Run("Given an already claimed id", func(t *testing.T) {
				t.Parallel()

				t.Run("Should keep the original caller id", func(t *testing.T) {
					t.Parallel()
					suite := WhenEnsuringConversationBeforeEach(tc)
					defer WhenEnsuringConversationAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					bob := strings.ToLower(common_tests.RandomEthAddress().Hex())

					created, err := suite.sut.Ensure(context.Background(), model_builders.NewConversationBuilder().
						WithId("conv-owned").
						WithAgentId(agentId).
						WithCallerId(alice).
						Build())
					assert.NoError(t, err)
					assert.True(t, created)

					created, err = suite.sut.Ensure(context.Background(), model_builders.NewConversationBuilder().
						WithId("conv-owned").
						WithAgentId(agentId).
						WithCallerId(bob).
						Build())
					assert.NoError(t, err)
					assert.False(t, created)

					stored, err := suite.sut.GetById(context.Background(), "conv-owned")
					assert.NoError(t, err)
					if assert.NotNil(t, stored) && assert.NotNil(t, stored.CallerId) {
						assert.Equal(t, alice, *stored.CallerId)
					}
				})
			})

			t.Run("Given a caller-less conversation is ensured again by a caller", func(t *testing.T) {
				t.Parallel()

				t.Run("Should keep it caller-less", func(t *testing.T) {
					t.Parallel()
					suite := WhenEnsuringConversationBeforeEach(tc)
					defer WhenEnsuringConversationAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())

					created, err := suite.sut.Ensure(context.Background(), model_builders.NewConversationBuilder().
						WithId("conv-console").
						WithAgentId(agentId).
						Build())
					assert.NoError(t, err)
					assert.True(t, created)

					created, err = suite.sut.Ensure(context.Background(), model_builders.NewConversationBuilder().
						WithId("conv-console").
						WithAgentId(agentId).
						WithCallerId(alice).
						Build())
					assert.NoError(t, err)
					assert.False(t, created)

					stored, err := suite.sut.GetById(context.Background(), "conv-console")
					assert.NoError(t, err)
					if assert.NotNil(t, stored) {
						assert.Nil(t, stored.CallerId)
					}
				})
			})
		})
	}
}
