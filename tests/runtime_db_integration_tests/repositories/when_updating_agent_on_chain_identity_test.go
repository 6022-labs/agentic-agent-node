package repositories_test

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenUpdatingAgentOnChainIdentityTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentRepository
}

func WhenUpdatingAgentOnChainIdentityBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenUpdatingAgentOnChainIdentityTestingSuite {
	databaseContainerBuilder := tc.ClientBuilderFn()
	dbName, connectionString := databaseContainerBuilder.Create()

	migrationFolderPath, err := common_tests.FindRuntimeMigrationFolderPath()
	if err != nil {
		panic(err)
	}

	dbClient, err := runtime_db.NewDatabaseClient(
		&settings.DatabaseSettings{Driver: tc.Driver, Url: connectionString},
		&settings.MigrationsSettings{FolderPath: *migrationFolderPath},
		zap.NewNop(),
	)
	if err != nil {
		panic(err)
	}
	if err := dbClient.Migrate(); err != nil {
		panic(err)
	}

	return &WhenUpdatingAgentOnChainIdentityTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentRepository(dbClient),
	}
}

func WhenUpdatingAgentOnChainIdentityAfterEach(suite *WhenUpdatingAgentOnChainIdentityTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingAgentOnChainIdentity(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no agent exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return ErrAgentNotInitialized", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingAgentOnChainIdentityBeforeEach(tc)
					defer WhenUpdatingAgentOnChainIdentityAfterEach(suite)

					err := suite.sut.UpdateOnChainIdentity(
						context.Background(),
						model_builders.NewAgentBuilder().Build(),
					)

					assert.ErrorIs(t, err, apperrors.ErrAgentNotInitialized)
				})

				t.Run("Should not create a row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingAgentOnChainIdentityBeforeEach(tc)
					defer WhenUpdatingAgentOnChainIdentityAfterEach(suite)

					_ = suite.sut.UpdateOnChainIdentity(context.Background(), model_builders.NewAgentBuilder().Build())

					got, err := suite.sut.Get(context.Background())
					assert.NoError(t, err)
					assert.Nil(t, got)
				})
			})

			t.Run("Given an agent already exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist the on-chain identity and soul", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingAgentOnChainIdentityBeforeEach(tc)
					defer WhenUpdatingAgentOnChainIdentityAfterEach(suite)

					inserted, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().Build())
					assert.NoError(t, err)

					owner := common_tests.RandomEthAddress().Hex()
					creator := common_tests.RandomEthAddress().Hex()
					err = suite.sut.UpdateOnChainIdentity(context.Background(), model_builders.NewAgentBuilder().
						WithId(inserted.Id).
						WithName("Babbage").
						WithCloneOf(7).
						WithCreator(creator).
						WithOwner(owner).
						WithImages(map[string]string{"avatar": "ipfs://new"}).
						WithAttributes(map[string]string{"role": "expert"}).
						WithSoul("new soul").
						Build())
					assert.NoError(t, err)

					reloaded, err := suite.sut.Get(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, "Babbage", lo.FromPtr(reloaded.Name))
					assert.Equal(t, uint64(7), lo.FromPtr(reloaded.CloneOf))
					assert.Equal(t, creator, lo.FromPtr(reloaded.Creator))
					assert.Equal(t, owner, lo.FromPtr(reloaded.Owner))
					assert.Equal(t, map[string]string{"avatar": "ipfs://new"}, reloaded.Images)
					assert.Equal(t, map[string]string{"role": "expert"}, reloaded.Attributes)
					assert.Equal(t, "new soul", lo.FromPtr(reloaded.Soul))
				})

				t.Run("Should keep its registration columns", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingAgentOnChainIdentityBeforeEach(tc)
					defer WhenUpdatingAgentOnChainIdentityAfterEach(suite)

					collectionAddress := common_tests.RandomEthAddress().Hex()
					inserted, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithChainId("80002").
						WithCollectionAddress(collectionAddress).
						WithCollectionTokenId(42).
						WithProposalId(3).
						WithEnsDomain("ada.agent.eth").
						Build())
					assert.NoError(t, err)

					err = suite.sut.UpdateOnChainIdentity(context.Background(), model_builders.NewAgentBuilder().
						WithId(inserted.Id).
						WithChainId("1").
						WithCollectionAddress(common_tests.RandomEthAddress().Hex()).
						WithCollectionTokenId(99).
						WithProposalId(9).
						WithEnsDomain("other.agent.eth").
						Build())
					assert.NoError(t, err)

					reloaded, err := suite.sut.Get(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, "80002", lo.FromPtr(reloaded.ChainId))
					assert.Equal(t, collectionAddress, reloaded.CollectionAddress)
					assert.Equal(t, uint64(42), lo.FromPtr(reloaded.CollectionTokenId))
					assert.Equal(t, uint64(3), lo.FromPtr(reloaded.ProposalId))
					assert.Equal(t, "ada.agent.eth", lo.FromPtr(reloaded.EnsDomain))
				})

				t.Run("Should bump UpdatedAt", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingAgentOnChainIdentityBeforeEach(tc)
					defer WhenUpdatingAgentOnChainIdentityAfterEach(suite)

					inserted, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().Build())
					assert.NoError(t, err)
					original, err := suite.sut.Get(context.Background())
					assert.NoError(t, err)

					time.Sleep(50 * time.Millisecond)

					err = suite.sut.UpdateOnChainIdentity(context.Background(), model_builders.NewAgentBuilder().
						WithId(inserted.Id).
						WithUpdatedAt(time.Unix(0, 0).UTC()).
						Build())
					assert.NoError(t, err)

					got, err := suite.sut.Get(context.Background())

					assert.NoError(t, err)
					if assert.NotNil(t, got) {
						assert.True(t, got.UpdatedAt.After(original.UpdatedAt))
					}
				})
			})
		})
	}
}
