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
	"fmt"
	"sort"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
)

// GetAccount gets the account data associated with a given address
func (c *Client) GetAccount(addr address.Address) (*backing.AccountData, error) {
	ads := make(map[string]*backing.AccountData)
	err := c.get(&ads, c.URL("account/account/%s", addr))
	if err != nil {
		return nil, err
	}
	ad, ok := ads[addr.String()]
	if !ok {
		return nil, nil
	}
	return ad, err
}

// GetAccount gets the account data associated with a given address
func GetAccount(node *Client, addr address.Address) (*backing.AccountData, error) {
	return node.GetAccount(addr)
}

// GetSequence gets the current sequence number of a particular account
func (c *Client) GetSequence(addr address.Address) (uint64, error) {
	ad, err := c.GetAccount(addr)
	if err != nil {
		// highest representable sequence number
		return ^uint64(0), err
	}
	if ad != nil {
		return ad.Sequence, nil
	}
	// accounts which don't exist have sequence 0
	return 0, nil
}

// GetSequence gets the current sequence number of a particular account
func GetSequence(node *Client, addr address.Address) (uint64, error) {
	return node.GetSequence(addr)
}

// GetAccountHistory gets account data history associated with a given address.
func (c *Client) GetAccountHistory(ahparams search.AccountHistoryParams) (*search.AccountHistoryResponse, error) {
	var response struct {
		Items []search.AccountTxValueData
	}
	err := c.get(&response, c.URLP(
		params{"after": ahparams.AfterHeight, "limit": ahparams.Limit},
		"account/history/%s", ahparams.Address))
	if err != nil {
		return nil, err
	}
	return &search.AccountHistoryResponse{
		Txs: response.Items,
	}, nil
}

// GetAccountHistory gets account data history associated with a given address.
func GetAccountHistory(node *Client, params search.AccountHistoryParams) (*search.AccountHistoryResponse, error) {
	return node.GetAccountHistory(params)
}

// GetAccountList gets a list of account names, paged according to the params
// Pass in after = "" (which is less than all nonempty strings) and limit = 0
// to get all results. (Note that the ndauapi will enforce a limit of 100 items.)
func (c *Client) GetAccountList(after string, limit int) (*query.AccountListQueryResponse, error) {
	resp := new(query.AccountListQueryResponse)
	err := c.get(resp, c.URLP(
		params{"after": after, "limit": limit},
		"account/list",
	))
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetAccountList gets a list of account names, paged according to the params
// Pass in after = "" (which is less than all nonempty strings) and limit = 0
// to get all results. (Note that the ndauapi will enforce a limit of 100 items.)
func GetAccountList(node *Client, after string, limit int) (*query.AccountListQueryResponse, error) {
	return node.GetAccountList(after, limit)
}

// GetAccountListBatch abstracts over the process of repeatedly calling
// GetAccountList in order to get a complete list of all known addresses.
//
// This function makes a best-effort attempt to return a complete and current
// list of accounts known to the node, but true consistency is impossible using
// a sequential paged API; as we cannot lock the node, there may be updates
// during paging which cause addresses to appear in pages we have already
// visited. This is unavoidable.
func (c *Client) GetAccountListBatch() ([]address.Address, error) {
	// nearly verbatim from
	// https://github.com/oneiro-ndev/ndau/blob/9198f7d7520854e68462de08d59daac93fe8a829/pkg/tool/account.go#L100-L150
	var (
		accts = make([]string, 0)
		after = ""
		limit = 100

		qaccts *query.AccountListQueryResponse
		err    error
	)

	getPage := func() {
		qaccts, err = c.GetAccountList(
			after,
			limit,
		)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf(
				"getPage(%s, %d)", after, limit,
			))
			return
		}
		accts = append(accts, qaccts.Accounts...)
		after = qaccts.NextAfter
	}

	// prime the pump
	getPage()
	if err != nil {
		return nil, err
	}
	for after != "" {
		getPage()
		if err != nil {
			return nil, err
		}
	}

	// eliminate duplicate accts and convert to address type
	sort.Strings(accts)
	addrs := make([]address.Address, 0, len(accts))
	for _, acct := range accts {
		addr, err := address.Validate(acct)
		if err != nil {
			return nil, errors.Wrap(err, "GetAccountListBatch validating acct addr")
		}
		if len(addrs) == 0 || addr != addrs[len(addrs)-1] {
			addrs = append(addrs, addr)
		}
	}

	return addrs, nil
}

// GetAccountListBatch abstracts over the process of repeatedly calling
// GetAccountList in order to get a complete list of all known addresses.
//
// This function makes a best-effort attempt to return a complete and current
// list of accounts known to the node, but true consistency is impossible using
// a sequential paged API; as we cannot lock the node, there may be updates
// during paging which cause addresses to appear in pages we have already
// visited. This is unavoidable.
func GetAccountListBatch(node *Client) ([]address.Address, error) {
	return node.GetAccountListBatch()
}
