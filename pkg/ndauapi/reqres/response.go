package reqres

// Statically validate Response implementation of Responder.
var _ Responder = (*Response)(nil)

// Response represents a response body and status code.
type Response struct {
	Bd  interface{}
	Sts int
}

// Status returns a status code and satisfies the Responder interface.
func (r Response) Status() int {
	return r.Sts
}

// Body returns a body interface and satisfies the Responder interface.
func (r Response) Body() interface{} {
	return r.Bd
}
