package search

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	metastate "github.com/ndau/metanode/pkg/meta/state"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndaumath/pkg/pricecurve"
)

// AppIndexable is an app which can help index its transactions.
//
// It's really only a thing in order to avoid circular imports; it will always
// in actuality be an ndau.App
type AppIndexable interface {
	GetAccountAddresses(tx metatx.Transactable) ([]string, error)
	GetState() metastate.State
	CalculateTxFeeNapu(tx metatx.Transactable) (uint64, error)
	CalculateTxSIBNapu(tx metatx.Transactable) (uint64, error)
}

// SysvarIndexable is a Transactable that has sysar data that we want to index.
type SysvarIndexable interface {
	metatx.Transactable

	// We use separate methods (instead of a struct to house the data) to avoid extra memory use.
	GetName() string
	GetValue() []byte
}

// MarketPriceIndexable is a Transactable that has updated the market price
type MarketPriceIndexable interface {
	metatx.Transactable

	GetMarketPrice() pricecurve.Nanocent
}

// TargetPriceIndexable is a marker interface for a Transactable, indicating
// that this tx has updated the target price.
type TargetPriceIndexable interface {
	metatx.Transactable

	UpdatedTargetPrice()
}
