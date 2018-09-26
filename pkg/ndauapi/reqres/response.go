package reqres

import "net/http"

// Statically validate Response implementation of Responder.
var _ Responder = (*Response)(nil)

// Response represents a response body and status code.
type Response struct {
	Bd  interface{}
	Sts int
}

// OKResponse builds a new response with status OK
func OKResponse(obj interface{}) Response {
	return Response{Bd: obj, Sts: http.StatusOK}
}

// Status returns a status code and satisfies the Responder interface.
func (r Response) Status() int {
	return r.Sts
}

// Body returns a body interface and satisfies the Responder interface.
func (r Response) Body() interface{} {
	return r.Bd
}
