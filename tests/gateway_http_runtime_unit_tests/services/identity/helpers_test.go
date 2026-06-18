package services_test

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
)

// httpTestFileHeader builds a *multipart.FileHeader from raw bytes for upload tests.
func httpTestFileHeader(t *testing.T, content []byte, contentType string) *multipart.FileHeader {
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

// readMultipartField parses the multipart body and returns a text field value,
// or a file part's filename when the field is a file.
func readMultipartField(t *testing.T, contentType string, reader io.Reader, field string) string {
	t.Helper()

	_, params, err := mime.ParseMediaType(contentType)
	assert.NoError(t, err)
	form, err := multipart.NewReader(reader, params["boundary"]).ReadForm(1 << 20)
	assert.NoError(t, err)

	if values := form.Value[field]; len(values) > 0 {
		return values[0]
	}
	if files := form.File[field]; len(files) > 0 {
		return files[0].Filename
	}
	return ""
}
