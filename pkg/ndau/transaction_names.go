package ndau

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"errors"
	"sort"
	"strings"

	metatx "github.com/ndau/metanode/pkg/meta/transaction"
)

var txnames map[string]metatx.Transactable

func init() {
	// initialize txnames map
	txnames = make(map[string]metatx.Transactable)
	// add all tx full names
	for _, example := range TxIDs {
		txnames[strings.ToLower(metatx.NameOf(example))] = example
	}
	// add common abbreviations
	txnames["rfe"] = TxIDs[3]                    // releasefromendowment
	txnames["crp"] = TxIDs[4]                    // changerecourseperiod
	txnames["change-recourse-period"] = TxIDs[4] // changerecourseperiod
	txnames["setv"] = TxIDs[10]                  // setvalidation
	txnames["set-validation"] = TxIDs[10]        // setvalidation
	txnames["nnr"] = TxIDs[13]                   // nominatenodereward
	txnames["cvc"] = TxIDs[16]                   // commandvalidatorchange
	txnames["create-child"] = TxIDs[21]          // createchildaccount
	txnames["create-child-account"] = TxIDs[21]  // createchildaccount
	txnames["record-price"] = TxIDs[22]          // recordprice
	txnames["ssv"] = TxIDs[23]                   // setsysvar

	//	remove obsolete abbreviations
	//	txnames["changesettlementperiod"] = TxIDs[4] // changesettlementperiod
	//	txnames["claim"] = TxIDs[10]                 // setvalidation
	//	txnames["claimaccount"] = TxIDs[10]          // setvalidation
	//	txnames["claim-child"] = TxIDs[21]           // createchildaccount
	//	txnames["claimchildaccount"] = TxIDs[21]     // createchildaccount
}

// KnownTxNames returns a list of valid names which can produce a tx
func KnownTxNames() []string {
	out := make([]string, 0, len(txnames))
	for n := range txnames {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// TxFromName prepares an example transaction from its name or from some known synonyms.
//
// This is most useful for generic unmarshaling.
func TxFromName(name string) (tx metatx.Transactable, err error) {
	example, ok := txnames[strings.ToLower(name)]
	if !ok {
		err = errors.New("Unknown transaction: " + name)
		return
	}

	tx = metatx.Clone(example)
	return
}
