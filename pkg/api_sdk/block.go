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
	"time"

	"github.com/pkg/errors"
	rpctypes "github.com/oneiro-ndev/tendermint.0.32.3/rpc/core/types"
)

// GetBlocksByHeight returns a sequence of block metadata for blocks with heights in the specified range.
//
// If noempty is set, exclude blocks containing no transactions
func (c *Client) GetBlocksByHeight(before, after uint64, noempty bool) (blocks *rpctypes.ResultBlockchainInfo, err error) {
	blocks = new(rpctypes.ResultBlockchainInfo)
	p := params{}
	if after > 0 {
		p["after"] = after
	}
	if noempty {
		p["filter"] = "noempty"
	}
	err = c.get(blocks, c.URLP(
		p,
		"block/before/%d", before,
	))
	err = errors.Wrap(err, "getting blocks by height")
	return
}

// GetBlockAt returns a block at a given height
//
// If height is 0, get the current block
func (c *Client) GetBlockAt(height uint64) (block *rpctypes.ResultBlock, err error) {
	block = new(rpctypes.ResultBlock)
	var url string
	if height == 0 {
		url = c.URL("block/current")
	} else {
		url = c.URL("block/height/%d", height)
	}
	err = c.get(block, url)
	err = errors.Wrap(err, "getting block by height")
	return
}

// GetCurrentBlock returns the current block
func (c *Client) GetCurrentBlock() (*rpctypes.ResultBlock, error) {
	return c.GetBlockAt(0)
}

// GetBlock returns a block with a particular hash
func (c *Client) GetBlock(hash string) (block *rpctypes.ResultBlock, err error) {
	block = new(rpctypes.ResultBlock)
	err = c.get(block, c.URL("block/hash/%s", hash))
	err = errors.Wrap(err, "getting block by hash")
	return
}

// GetBlocksByRange returns a sequence of block metadata for blocks with heights in the specified range.
//
// If noempty is set, exclude blocks containing no transactions
func (c *Client) GetBlocksByRange(before, after uint64, noempty bool) (blocks *rpctypes.ResultBlockchainInfo, err error) {
	blocks = new(rpctypes.ResultBlockchainInfo)
	p := params{}
	if noempty {
		p["noempty"] = "true"
	}
	err = c.get(blocks, c.URLP(
		p,
		"block/range/%d/%d", after, before,
	))
	err = errors.Wrap(err, "getting blocks by height range")
	return
}

// GetBlocksByDaterange returns a sequence of block metadata for blocks with block times in the specified range.
//
// If noempty is set, exclude blocks containing no transactions.
// after should be the last value from the previous page, or a zero value to exclude
func (c *Client) GetBlocksByDaterange(first, last time.Time, noempty bool, after time.Time, limit int) (blocks *rpctypes.ResultBlockchainInfo, err error) {
	blocks = new(rpctypes.ResultBlockchainInfo)
	p := params{}
	if after != (time.Time{}) {
		p["after"] = after.Format(time.RFC3339)
	}
	if limit != 0 {
		p["limit"] = limit
	}
	err = c.get(
		blocks,
		c.URLP(
			p,
			"block/daterange/%s/%s", first.Format(time.RFC3339), last.Format(time.RFC3339),
		),
	)
	err = errors.Wrap(err, "getting blocks by date range")
	return
}
