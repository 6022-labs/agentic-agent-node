package requests_test

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	common_tests "github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSaveAgentRegistrationRequest(t *testing.T) {
	t.Parallel()

	generateSaveAgentRegistrationRequest := func(t *testing.T) *requests.SaveAgentRegistrationRequest {
		return &requests.SaveAgentRegistrationRequest{
			ChainIdRequest:    requests.ChainIdRequest{ChainId: "80002"},
			CollectionAddress: "  " + common_tests.RandomEthAddress().Hex() + "  ",
			Owner:             common_tests.RandomEthAddress().Hex(),
			Name:              "my-agent",
			AttributesRaw:     `{"role":"expert","specialty":"defi"}`,
			Image:             newImageFileHeader(t, []byte("fake-image")),
		}
	}

	t.Run("Given a blank chain id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it on the chainId field", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.ChainId = "   "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "chainId")
		})
	})

	t.Run("Given a malformed collection address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it on the collectionAddress field", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.CollectionAddress = "not-an-address"

			assert.ErrorContains(t, req.ValidateAndSanitize(), "collectionAddress")
		})
	})

	t.Run("Given a malformed owner", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it on the owner field", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.Owner = "0x123"

			assert.ErrorContains(t, req.ValidateAndSanitize(), "owner")
		})
	})

	t.Run("Given a blank name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it on the name field", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.Name = "   "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a malformed ENS label", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it on the name field", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"my_agent", "foo--bar", "-foo", "foo-"} {
				req := generateSaveAgentRegistrationRequest(t)
				req.Name = name

				assert.ErrorContains(t, req.ValidateAndSanitize(), "name", name)
			}
		})
	})

	t.Run("Given an image larger than 5MB", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it on the image field", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.Image.Size = 5*1024*1024 + 1

			assert.ErrorContains(t, req.ValidateAndSanitize(), "image")
		})
	})

	t.Run("Given malformed attributes", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it on the attributes field", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.AttributesRaw = `{"specialty": 42}`

			assert.ErrorContains(t, req.ValidateAndSanitize(), "attributes")
		})
	})

	t.Run("Given the attributes carry no role", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it on the attributes field", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.AttributesRaw = `{"specialty":"defi"}`

			assert.ErrorContains(t, req.ValidateAndSanitize(), "attributes")
		})
	})

	t.Run("Given a blank role attribute", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it on the attributes field", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.AttributesRaw = `{"role":"  "}`

			assert.ErrorContains(t, req.ValidateAndSanitize(), "attributes")
		})
	})

	t.Run("Given a free-form role attribute", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.AttributesRaw = `{"role":"wizard"}`

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given no image", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.Image = nil

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sanitize the fields", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "my-agent", req.Name)
			assert.NotContains(t, req.CollectionAddress, " ")
		})

		t.Run("Should parse the attributes", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, map[string]string{"role": "expert", "specialty": "defi"}, req.Attributes)
		})
	})

}

// newImageFileHeader builds a *multipart.FileHeader carrying raw bytes.
func newImageFileHeader(t *testing.T, content []byte) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"; filename="base.png"`)
	h.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(h)
	assert.NoError(t, err)
	_, err = part.Write(content)
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	form, err := multipart.NewReader(body, writer.Boundary()).ReadForm(1 << 20)
	assert.NoError(t, err)
	return form.File["image"][0]
}
