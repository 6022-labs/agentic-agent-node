package models

import "time"

type ExternalSelfToolPayload struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Group       string    `json:"group"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ExternalSelfToolListPayload struct {
	Tools []ExternalSelfToolPayload `json:"tools"`
}

// ExternalUpdateSelfToolPayload is the body shape /internal/config/self-tools/:id PUT accepts.
type ExternalUpdateSelfToolPayload struct {
	Enabled bool `json:"enabled"`
}
