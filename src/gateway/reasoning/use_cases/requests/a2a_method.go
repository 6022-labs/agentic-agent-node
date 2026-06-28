package requests

// A2AMethod is the JSON-RPC method of an A2A request; only SendMessage is handled.
type A2AMethod string

const (
	A2AMethodSendMessage A2AMethod = "SendMessage"
)

func (m A2AMethod) String() string {
	return string(m)
}
