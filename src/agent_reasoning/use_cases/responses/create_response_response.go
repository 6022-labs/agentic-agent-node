package responses

import "time"

type CreateResponseMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type CreateResponseResponse struct {
	Message *CreateResponseMessage `json:"message,omitempty"`
}
