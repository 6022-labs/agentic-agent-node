package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strconv"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpAgentRegistrationClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpAgentRegistrationClient(client clients.ClientWithResponsesInterface) *HttpAgentRegistrationClient {
	return &HttpAgentRegistrationClient{client: client}
}

func (c *HttpAgentRegistrationClient) Save(ctx context.Context, body *models.ExternalSaveAgentRegistrationPayload) (*models.ExternalAgentRegistrationPayload, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if err := c.writeSaveFields(writer, body); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	if body.Image != nil {
		if err := c.writeImagePart(writer, body.Image); err != nil {
			return nil, helpers.MapRuntimeError(err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}

	resp, err := c.client.PostInternalAgentRegistrationWithBodyWithResponse(ctx, writer.FormDataContentType(), &buf)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalAgentRegistrationPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

// writeSaveFields writes the non-file draft fields; attributes go as a JSON object string.
func (c *HttpAgentRegistrationClient) writeSaveFields(writer *multipart.Writer, body *models.ExternalSaveAgentRegistrationPayload) error {
	fields := map[string]string{
		"collectionAddress": body.CollectionAddress,
		"owner":             body.Owner,
		"name":              body.Name,
		"role":              body.Role,
	}
	if body.CloneOf != nil {
		fields["cloneOf"] = strconv.FormatUint(*body.CloneOf, 10)
	}
	if len(body.Attributes) > 0 {
		attrs, err := json.Marshal(body.Attributes)
		if err != nil {
			return err
		}
		fields["attributes"] = string(attrs)
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return err
		}
	}
	return nil
}

// writeImagePart streams the uploaded file, preserving its content type so the runtime can read the mime.
func (c *HttpAgentRegistrationClient) writeImagePart(writer *multipart.Writer, fileHeader *multipart.FileHeader) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename=%q`, fileHeader.Filename))
	h.Set("Content-Type", contentType)

	part, err := writer.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, src)
	return err
}

func (c *HttpAgentRegistrationClient) Get(ctx context.Context) (*models.ExternalAgentRegistrationPayload, error) {
	resp, err := c.client.GetInternalAgentRegistrationWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalAgentRegistrationPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) MintInfo(ctx context.Context) (*models.ExternalMintInfoPayload, error) {
	resp, err := c.client.GetInternalAgentRegistrationMintInfoWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalMintInfoPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) CheckMint(ctx context.Context) (*models.ExternalRegistrationStatusPayload, error) {
	resp, err := c.client.PostInternalAgentRegistrationMintCheckWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalRegistrationStatusPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) EnsInfo(ctx context.Context) (*models.ExternalEnsInfoPayload, error) {
	resp, err := c.client.GetInternalAgentRegistrationEnsInfoWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalEnsInfoPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) CheckEns(ctx context.Context) (*models.ExternalRegistrationStatusPayload, error) {
	resp, err := c.client.PostInternalAgentRegistrationEnsCheckWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalRegistrationStatusPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) Self(ctx context.Context) (*models.ExternalRegistrationStatusPayload, error) {
	resp, err := c.client.PostInternalAgentRegistrationSelfWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalRegistrationStatusPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}
