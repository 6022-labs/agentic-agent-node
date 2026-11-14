package services_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const (
	syncWaitTimeout = 2 * time.Second
	noSyncWindow    = 100 * time.Millisecond
)

type WhenResolvingAgentTestingSuite struct {
	sut *services.AgentResolver

	logs                       *observer.ObservedLogs
	mockAgentRepository        *repository_mocks.MockAgentRepositoryInterface
	mockRegistrationReconciler *interface_mocks.MockRegistrationReconcilerInterface
}

func WhenResolvingAgentBeforeEach(t *testing.T, identitySyncInterval time.Duration) *WhenResolvingAgentTestingSuite {
	mockController := gomock.NewController(t)

	core, logs := observer.New(zapcore.WarnLevel)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockRegistrationReconciler := interface_mocks.NewMockRegistrationReconcilerInterface(mockController)

	sut := services.NewAgentResolver(
		zap.New(core),
		mockAgentRepository,
		mockRegistrationReconciler,
		&settings.AgentIdentitySettings{IdentitySyncInterval: identitySyncInterval},
	)

	return &WhenResolvingAgentTestingSuite{
		sut:                        sut,
		logs:                       logs,
		mockAgentRepository:        mockAgentRepository,
		mockRegistrationReconciler: mockRegistrationReconciler,
	}
}

func TestWhenResolvingAgent(t *testing.T) {
	t.Parallel()

	waitFor := func(t *testing.T, done <-chan struct{}) {
		t.Helper()
		select {
		case <-done:
		case <-time.After(syncWaitTimeout):
			t.Fatal("timed out")
		}
	}
	// countSyncs lets a late, unexpected background sync be observed instead of racing the test end.
	countSyncs := func(suite *WhenResolvingAgentTestingSuite, err error) *atomic.Int32 {
		var syncs atomic.Int32
		suite.mockRegistrationReconciler.EXPECT().
			ReconcileIdentity(gomock.Any()).
			DoAndReturn(func(context.Context) error {
				syncs.Add(1)
				return err
			}).
			AnyTimes()
		return &syncs
	}

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		lookupErr := errors.New("db down")

		initSuite := func(suite *WhenResolvingAgentTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, lookupErr)
		}

		t.Run("Should propagate the error without syncing", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentBeforeEach(t, time.Hour)
			initSuite(suite)
			syncs := countSyncs(suite, nil)

			agent, err := suite.sut.Get(context.Background())

			assert.Nil(t, agent)
			assert.ErrorIs(t, err, lookupErr)
			assert.Never(t, func() bool { return syncs.Load() > 0 }, noSyncWindow, time.Millisecond)
		})
	})

	t.Run("Given no agent row exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenResolvingAgentTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return nil without syncing", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentBeforeEach(t, time.Hour)
			initSuite(suite)
			syncs := countSyncs(suite, nil)

			agent, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.Nil(t, agent)
			assert.Never(t, func() bool { return syncs.Load() > 0 }, noSyncWindow, time.Millisecond)
		})
	})

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		unminted := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()

		initSuite := func(suite *WhenResolvingAgentTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(unminted, nil)
		}

		t.Run("Should return the agent without syncing", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentBeforeEach(t, time.Hour)
			initSuite(suite)
			syncs := countSyncs(suite, nil)

			agent, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.Same(t, unminted, agent)
			assert.Never(t, func() bool { return syncs.Load() > 0 }, noSyncWindow, time.Millisecond)
		})
	})

	t.Run("Given a minted agent whose identity is stale", func(t *testing.T) {
		t.Parallel()

		minted := model_builders.NewAgentBuilder().WithCollectionTokenId(5).Build()

		initSuite := func(suite *WhenResolvingAgentTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(minted, nil).AnyTimes()
		}

		t.Run("Should return the agent before the sync finishes", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentBeforeEach(t, time.Hour)
			initSuite(suite)
			started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
			suite.mockRegistrationReconciler.EXPECT().
				ReconcileIdentity(gomock.Any()).
				DoAndReturn(func(context.Context) error {
					defer close(done)
					close(started)
					<-release
					return nil
				})

			returned := make(chan struct{})
			go func() {
				defer close(returned)
				agent, err := suite.sut.Get(context.Background())
				assert.NoError(t, err)
				assert.Same(t, minted, agent)
			}()

			waitFor(t, started)
			waitFor(t, returned)
			close(release)
			waitFor(t, done)
		})

		t.Run("Should sync with a bounded context detached from the request", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentBeforeEach(t, time.Hour)
			initSuite(suite)
			syncCtxErr, syncCtxHasDeadline := make(chan error, 1), make(chan bool, 1)
			suite.mockRegistrationReconciler.EXPECT().
				ReconcileIdentity(gomock.Any()).
				DoAndReturn(func(ctx context.Context) error {
					_, hasDeadline := ctx.Deadline()
					syncCtxHasDeadline <- hasDeadline
					syncCtxErr <- ctx.Err()
					return nil
				})
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := suite.sut.Get(ctx)

			assert.NoError(t, err)
			select {
			case err := <-syncCtxErr:
				assert.NoError(t, err)
				assert.True(t, <-syncCtxHasDeadline)
			case <-time.After(syncWaitTimeout):
				t.Fatal("timed out")
			}
		})

		t.Run("Should sync once for concurrent calls", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentBeforeEach(t, time.Hour)
			initSuite(suite)
			var syncs atomic.Int32
			release := make(chan struct{})
			suite.mockRegistrationReconciler.EXPECT().
				ReconcileIdentity(gomock.Any()).
				DoAndReturn(func(context.Context) error {
					syncs.Add(1)
					<-release
					return nil
				}).
				AnyTimes()

			var wg sync.WaitGroup
			for range 10 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, _ = suite.sut.Get(context.Background())
				}()
			}
			wg.Wait()
			close(release)

			assert.Eventually(t, func() bool { return syncs.Load() == 1 }, syncWaitTimeout, time.Millisecond)
			assert.Never(t, func() bool { return syncs.Load() > 1 }, noSyncWindow, time.Millisecond)
		})

		t.Run("Should not sync again within the interval", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentBeforeEach(t, time.Hour)
			initSuite(suite)
			syncs := countSyncs(suite, nil)

			_, _ = suite.sut.Get(context.Background())
			assert.Eventually(t, func() bool { return syncs.Load() == 1 }, syncWaitTimeout, time.Millisecond)
			// Lets the first sync release its in-flight flag, so only the interval can block the next one.
			time.Sleep(noSyncWindow)
			_, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.Never(t, func() bool { return syncs.Load() > 1 }, noSyncWindow, time.Millisecond)
		})
	})

	t.Run("Given the sync interval has elapsed", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenResolvingAgentTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(model_builders.NewAgentBuilder().WithCollectionTokenId(5).Build(), nil).
				AnyTimes()
		}

		t.Run("Should sync again on the next call", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentBeforeEach(t, time.Nanosecond)
			initSuite(suite)
			syncs := countSyncs(suite, nil)

			assert.Eventually(t, func() bool {
				_, _ = suite.sut.Get(context.Background())
				return syncs.Load() >= 2
			}, syncWaitTimeout, time.Millisecond)
		})
	})

	t.Run("Given the sync fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenResolvingAgentTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(model_builders.NewAgentBuilder().WithCollectionTokenId(5).Build(), nil).
				AnyTimes()
		}

		t.Run("Should log a warning", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentBeforeEach(t, time.Hour)
			initSuite(suite)
			countSyncs(suite, errors.New("rpc"))

			_, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.Eventually(t, func() bool {
				return suite.logs.FilterMessage("agent identity sync failed").Len() == 1
			}, syncWaitTimeout, time.Millisecond)
		})

		t.Run("Should not retry within the interval", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentBeforeEach(t, time.Hour)
			initSuite(suite)
			syncs := countSyncs(suite, errors.New("rpc"))

			_, _ = suite.sut.Get(context.Background())
			assert.Eventually(t, func() bool {
				return suite.logs.FilterMessage("agent identity sync failed").Len() == 1
			}, syncWaitTimeout, time.Millisecond)
			time.Sleep(noSyncWindow)
			_, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.Never(t, func() bool { return syncs.Load() > 1 }, noSyncWindow, time.Millisecond)
		})
	})
}
