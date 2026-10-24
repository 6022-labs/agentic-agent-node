package middlewares_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares"
	middleware_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares/interfaces"
	core_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/core_mocks/services_mocks/interfaces_mocks"
	access_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mvc_mocks/reasoning_mocks/middlewares_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const jwtTestToken = "header.payload.signature"

type WhenGrantingJwtFreeAccessTestingSuite struct {
	app              *fiber.App
	mockJwtValidator *core_mocks.MockJwtValidatorInterface
	matcherA         *access_mocks.MockFreeRuleMatcherInterface
	matcherB         *access_mocks.MockFreeRuleMatcherInterface

	downstreamRan bool
	seenAddress   string
}

func WhenGrantingJwtFreeAccessBeforeEach(t *testing.T) *WhenGrantingJwtFreeAccessTestingSuite {
	mockController := gomock.NewController(t)
	return &WhenGrantingJwtFreeAccessTestingSuite{
		mockJwtValidator: core_mocks.NewMockJwtValidatorInterface(mockController),
		matcherA:         access_mocks.NewMockFreeRuleMatcherInterface(mockController),
		matcherB:         access_mocks.NewMockFreeRuleMatcherInterface(mockController),
	}
}

// build wires the link over matchers; a deferral answers 418, so downstream runs only on a grant.
func (s *WhenGrantingJwtFreeAccessTestingSuite) build(matchers ...middleware_interfaces.FreeRuleMatcherInterface) {
	sut := middlewares.NewJwtFreeAccessMethod(middlewares.NewJwtFreeAccessMethodParams{
		JwtValidator:     s.mockJwtValidator,
		FreeRuleMatchers: matchers,
	})
	app := fiber.New()
	app.Post("/responses",
		func(ctx fiber.Ctx) error {
			handled, err := sut.TryHandle(ctx, nil)
			if handled {
				return err
			}
			return ctx.SendStatus(fiber.StatusTeapot)
		},
		func(ctx fiber.Ctx) error {
			s.downstreamRan = true
			s.seenAddress, _ = helpers.CallerIdFromContext(ctx.Context())
			return ctx.JSON(fiber.Map{"ok": true})
		},
	)
	s.app = app
}

func jwtAuthRequest(authorization string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/responses", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	return req
}

func TestWhenGrantingJwtFreeAccess(t *testing.T) {
	t.Parallel()

	t.Run("Given no Authorization header", func(t *testing.T) {
		t.Parallel()

		t.Run("Should defer without touching the validator", func(t *testing.T) {
			t.Parallel()

			suite := WhenGrantingJwtFreeAccessBeforeEach(t)
			suite.build(suite.matcherA)

			resp, err := suite.app.Test(jwtAuthRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusTeapot, resp.StatusCode)
			assert.False(t, suite.downstreamRan)
		})
	})

	t.Run("Given a non-Bearer scheme", func(t *testing.T) {
		t.Parallel()

		t.Run("Should defer", func(t *testing.T) {
			t.Parallel()

			suite := WhenGrantingJwtFreeAccessBeforeEach(t)
			suite.build(suite.matcherA)

			resp, err := suite.app.Test(jwtAuthRequest("Token " + jwtTestToken))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusTeapot, resp.StatusCode)
		})
	})

	t.Run("Given a Bearer scheme with a blank token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should defer without touching the validator", func(t *testing.T) {
			t.Parallel()

			suite := WhenGrantingJwtFreeAccessBeforeEach(t)
			suite.build(suite.matcherA)

			resp, err := suite.app.Test(jwtAuthRequest("Bearer "))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusTeapot, resp.StatusCode)
		})
	})

	t.Run("Given the token fails validation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should defer without consulting the matchers", func(t *testing.T) {
			t.Parallel()

			suite := WhenGrantingJwtFreeAccessBeforeEach(t)
			suite.mockJwtValidator.EXPECT().
				Verify(jwtTestToken).
				Return(common.Address{}, errors.New("expired")).
				Times(1)
			suite.build(suite.matcherA)

			resp, err := suite.app.Test(jwtAuthRequest("Bearer " + jwtTestToken))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusTeapot, resp.StatusCode)
		})
	})

	t.Run("Given a valid token but no matcher grants free access", func(t *testing.T) {
		t.Parallel()

		t.Run("Should defer", func(t *testing.T) {
			t.Parallel()

			suite := WhenGrantingJwtFreeAccessBeforeEach(t)
			suite.mockJwtValidator.EXPECT().
				Verify(jwtTestToken).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
			suite.matcherA.EXPECT().
				MatchesFreeRule(gomock.Any(), common.HexToAddress(x402TestPayerHex)).
				Return(false).
				Times(1)
			suite.build(suite.matcherA)

			resp, err := suite.app.Test(jwtAuthRequest("Bearer " + jwtTestToken))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusTeapot, resp.StatusCode)
			assert.False(t, suite.downstreamRan)
		})
	})

	t.Run("Given a valid token the matcher deems free-entitled", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGrantingJwtFreeAccessTestingSuite) {
			suite.mockJwtValidator.EXPECT().
				Verify(jwtTestToken).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
			suite.matcherA.EXPECT().
				MatchesFreeRule(gomock.Any(), common.HexToAddress(x402TestPayerHex)).
				Return(true).
				Times(1)
			suite.build(suite.matcherA)
		}

		t.Run("Should attach the caller id to the context", func(t *testing.T) {
			t.Parallel()

			suite := WhenGrantingJwtFreeAccessBeforeEach(t)
			initSuite(suite)

			_, err := suite.app.Test(jwtAuthRequest("Bearer " + jwtTestToken))

			assert.NoError(t, err)
			assert.Equal(t, strings.ToLower(x402TestPayerHex), suite.seenAddress)
		})

		t.Run("Should run the downstream handler", func(t *testing.T) {
			t.Parallel()

			suite := WhenGrantingJwtFreeAccessBeforeEach(t)
			initSuite(suite)

			resp, err := suite.app.Test(jwtAuthRequest("Bearer " + jwtTestToken))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.True(t, suite.downstreamRan)
		})
	})

	t.Run("Given the first matcher declines and the second grants", func(t *testing.T) {
		t.Parallel()

		t.Run("Should grant access", func(t *testing.T) {
			t.Parallel()

			suite := WhenGrantingJwtFreeAccessBeforeEach(t)
			suite.mockJwtValidator.EXPECT().
				Verify(jwtTestToken).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
			gomock.InOrder(
				suite.matcherA.EXPECT().MatchesFreeRule(gomock.Any(), gomock.Any()).Return(false),
				suite.matcherB.EXPECT().MatchesFreeRule(gomock.Any(), gomock.Any()).Return(true),
			)
			suite.build(suite.matcherA, suite.matcherB)

			resp, err := suite.app.Test(jwtAuthRequest("Bearer " + jwtTestToken))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.True(t, suite.downstreamRan)
		})
	})
}
