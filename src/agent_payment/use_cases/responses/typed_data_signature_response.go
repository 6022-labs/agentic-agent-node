package responses

type TypedDataSignatureResponse struct {
	Signature string `json:"signature"`
	Address   string `json:"address"`
}
