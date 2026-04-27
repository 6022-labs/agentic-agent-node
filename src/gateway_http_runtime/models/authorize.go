package models

type AuthorizePayload struct {
	Address string `json:"address"`
}

type AuthorizeResultPayload struct {
	Authorized bool `json:"authorized"`
}
