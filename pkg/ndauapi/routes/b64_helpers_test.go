package routes_test

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"encoding/base64"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
)

func b64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func b64str(s string) string {
	return b64([]byte(s))
}

func b64Tx(tx ndau.NTransactable) (string, error) {
	m, err := metatx.Marshal(tx, ndau.TxIDs)
	if err != nil {
		return "", err
	}
	return b64(m), nil
}
