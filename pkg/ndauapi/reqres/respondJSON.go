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
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// When no code is specified.
const defaultStatusCode = http.StatusInternalServerError

// Responder provides a status code.
type Responder interface {
	Status() int
	Body() interface{}
}

// RespondJSON marshals and writes to the api user and logs.
func RespondJSON(w http.ResponseWriter, res Responder) {
	status := res.Status()
	if status == 0 {
		status = defaultStatusCode // Default if nothing specified.
	}

	body := res.Body()
	resB, err := json.Marshal(body)
	if err != nil {
		log.Errorf("could not marshal response body to json: %v", err)
		// Attempt to respond with an error message.
		status = http.StatusInternalServerError
		resB, _ = json.Marshal(ErrorBody{Message: "Could not encode response."}) // ignores the error, already in an error state.
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("X-Robots-Tag", "noindex")
	w.WriteHeader(status)  // doesn't return an error
	_, err = w.Write(resB) // ignoring bytes written
	if err != nil {
		log.Errorf("could not write response: %v", err)
	}
}
