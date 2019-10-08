package config

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"encoding"
	"encoding/base64"
)

// what we _really_ want is for this type to come from chaincode itself,
// but for now we can get away with stubbing it out
type chaincode []byte

func (c chaincode) MarshalText() (text []byte, err error) {
	text = make([]byte, base64.StdEncoding.EncodedLen(len(c)))
	base64.StdEncoding.Encode(text, c)
	return
}

var _ encoding.TextMarshaler = (chaincode)(nil)

func (c chaincode) UnmarshalText(text []byte) (err error) {
	c = make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	_, err = base64.StdEncoding.Decode(c, text)
	return
}

var _ encoding.TextUnmarshaler = (chaincode)(nil)
