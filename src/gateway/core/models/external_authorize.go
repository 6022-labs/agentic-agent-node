package models

type ExternalAuthorizePayload struct {
	Address string `json:"address"`
}

type ExternalAuthorizeResultPayload struct {
	Authorized bool `json:"authorized"`
}
