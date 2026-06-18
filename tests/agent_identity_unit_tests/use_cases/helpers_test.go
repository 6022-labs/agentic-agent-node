package use_cases_test

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
)

// newMultipartFileHeader builds a *multipart.FileHeader from raw bytes so tests
// can feed the save use case an uploaded image.
func newMultipartFileHeader(t *testing.T, content []byte, contentType string) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"; filename="base.png"`)
	h.Set("Content-Type", contentType)
	part, err := writer.CreatePart(h)
	assert.NoError(t, err)
	_, err = part.Write(content)
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	form, err := multipart.NewReader(body, writer.Boundary()).ReadForm(1 << 20)
	assert.NoError(t, err)
	return form.File["image"][0]
}
