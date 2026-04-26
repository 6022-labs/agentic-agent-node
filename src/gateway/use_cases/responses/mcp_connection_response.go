package responses

type McpConnectionResponse struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Url            string `json:"url"`
	AuthHeaderName string `json:"authHeaderName,omitempty"`
}
