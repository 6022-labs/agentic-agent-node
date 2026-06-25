package interfaces

// Span is a started span the caller must End; Fail marks it errored.
type Span interface {
	Fail(err error)
	End()
}
