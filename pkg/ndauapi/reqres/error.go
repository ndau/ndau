package reqres

// Statically validate Error implementation of Responder.
var _ Responder = (*Error)(nil)

// ErrorBody represents a simple error message
type ErrorBody struct {
	Message string `json:"msg"`
}

// Error represents an error with a message and an http status code.
type Error struct {
	ErrorBody ErrorBody
	Sts       int
}

// NewError returns an Error that RespondJSON can use.
func NewError(message string, status int) Error {
	return Error{
		ErrorBody: ErrorBody{
			Message: message,
		},
		Sts: status,
	}
}

// Status returns a status code and satisfies the Responder interface.
func (e Error) Status() int {
	return e.Sts
}

// Body returns a status code and satisfies the Responder interface.
func (e Error) Body() interface{} {
	return e.ErrorBody
}
