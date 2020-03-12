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
	"fmt"

	"github.com/ndau/ndaumath/pkg/signature"
)

// A Keypair holds a pair of keys
type Keypair struct {
	Path    *string              `toml:"path"`
	Public  signature.PublicKey  `toml:"public"`
	Private signature.PrivateKey `toml:"private"`
}

// String satisfies io.Stringer by writing the public key
func (kp Keypair) String() string {
	if kp.Path == nil {
		return fmt.Sprintf("<keypair: pub %s>", kp.Public.String())
	}
	return fmt.Sprintf(
		"<keypair %s: pub %s>",
		kp.Public.String(),
		*kp.Path,
	)
}
