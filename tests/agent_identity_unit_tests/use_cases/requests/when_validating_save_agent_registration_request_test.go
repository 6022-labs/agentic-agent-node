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
			ChainId:           "80002",
			CollectionAddress: "  " + common_tests.RandomEthAddress().Hex() + "  ",
			Owner:             common_tests.RandomEthAddress().Hex(),
			Name:              "my-agent",
			AttributesRaw:     `{"role":"expert","specialty":"defi"}`,
			Image:             newImageFileHeader(t, []byte("fake-image")),
		}
	}

	t.Run("Given the request is valid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sanitize the fields and parse the attributes", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			err := req.ValidateAndSanitize()

			assert.NoError(t, err)
			assert.Equal(t, "my-agent", req.Name)
			assert.NotContains(t, req.CollectionAddress, " ")
			assert.Equal(t, map[string]string{"role": "expert", "specialty": "defi"}, req.Attributes)
		})
	})

	t.Run("Given the attributes are absent", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error for the missing role", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.AttributesRaw = "   "

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the chain id is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.ChainId = "   "

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the collection address is not a hex address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.CollectionAddress = "not-an-address"

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the owner is not a hex address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.Owner = "0x123"

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the name is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.Name = "   "

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the name is not a valid ENS label", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"my_agent", "foo--bar", "-foo", "foo-"} {
				req := generateSaveAgentRegistrationRequest(t)
				req.Name = name

				assert.Error(t, req.ValidateAndSanitize(), name)
			}
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

	t.Run("Given the attributes carry no role", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.AttributesRaw = `{"specialty":"defi"}`

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the role attribute is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.AttributesRaw = `{"role":"  "}`

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the image is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.Image = nil

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the image exceeds 5MB", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.Image.Size = 5*1024*1024 + 1

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the attributes are not a JSON object of strings", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSaveAgentRegistrationRequest(t)
			req.AttributesRaw = `{"specialty": 42}`

			assert.Error(t, req.ValidateAndSanitize())
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
