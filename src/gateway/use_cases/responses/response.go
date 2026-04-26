package responses

import "time"

type OutputMessage struct {
	Id        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ResponseStats struct {
	Iterations  int `json:"iterations"`
	ToolCalls   int `json:"toolCalls"`
	Tokens      int `json:"tokens"`
	WallClockMs int `json:"wallClockMs"`
}

type Response struct {
	Id      string           `json:"id"`
	IsFinal bool             `json:"isFinal"`
	Output  []*OutputMessage `json:"output"`
	Stats   *ResponseStats   `json:"stats,omitempty"`
}
