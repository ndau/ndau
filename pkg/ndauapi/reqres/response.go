package reqres

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

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
