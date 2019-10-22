package reqres

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"
	"strings"
)

// Statically validate APIError implementation of Responder.
var _ Responder = (*APIError)(nil)

// ErrorBody represents a simple error message
type ErrorBody struct {
	Message string `json:"msg"`
	Log     string `json:"log,omitempty"`
}

// APIError represents an error with a message and an http status code.
type APIError struct {
	ErrorBody ErrorBody
	Sts       int
}

// NewAPIError returns an APIError that RespondJSON can use.
func NewAPIError(message string, status int) APIError {
	return APIError{
		ErrorBody: ErrorBody{
			Message: message,
		},
		Sts: status,
	}
}

// NewFromErr builds an APIError from a go error
func NewFromErr(msg string, err error, status int, logs ...string) APIError {
	logcat := strings.Join(logs, ", ")
	return APIError{
		ErrorBody: ErrorBody{
			Message: fmt.Sprintf("%s (%v)", msg, err),
			Log:     logcat,
		},
		Sts: status,
	}
}

// Status returns a status code and satisfies the Responder interface.
func (e APIError) Status() int {
	return e.Sts
}

// Body returns a status code and satisfies the Responder interface.
func (e APIError) Body() interface{} {
	return e.ErrorBody
}
