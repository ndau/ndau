package sdk

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
)

// GetCurrencySeats gets a list of ndau currency seats
func (c *Client) GetCurrencySeats() (seats []address.Address, err error) {
	seats = make([]address.Address, 0)
	err = c.get(&seats, c.URL("account/currencyseats"))
	err = errors.Wrap(err, "getting currency seats from API")
	return
}

// GetCurrencySeats gets a list of ndau currency seats
func GetCurrencySeats(node *Client) ([]address.Address, error) {
	return node.GetCurrencySeats()
}

// GetDelegates gets the set of nodes with delegates, and the list of accounts delegated to each
func (c *Client) GetDelegates() (delegates map[address.Address][]address.Address, err error) {
	delegates = make(map[address.Address][]address.Address)
	err = c.get(&delegates, c.URL("state/delegates"))
	err = errors.Wrap(err, "getting delegates from API")
	return
}

// GetDelegates gets the set of nodes with delegates, and the list of accounts delegated to each
func GetDelegates(node *Client) (map[address.Address][]address.Address, error) {
	return node.GetDelegates()
}
