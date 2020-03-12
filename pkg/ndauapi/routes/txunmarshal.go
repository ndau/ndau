package routes

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
	"io"
	"io/ioutil"

	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau"
)

// TxNames returns a list of all of the valid transaction names, plus some synonyms.
func TxNames() []string {
	return ndau.KnownTxNames()
}

// TxUnmarshal constructs an object containing transaction data.
// Given the name of a transaction type and a reader containing the JSON for a transaction
// (usually the request Body from a POST), this constructs a new object containing that
// transactions's data.
func TxUnmarshal(txtype string, r io.Reader) (metatx.Transactable, error) {
	tx, err := ndau.TxFromName(txtype)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buf, tx)
	return tx, err
}
