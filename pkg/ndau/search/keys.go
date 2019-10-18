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
	"fmt"
	"strings"
	"time"

	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// We use these prefixes to help us group keys in the index.  They could prove useful if we ever
// want to do things like "wipe all hash-to-height keys" without affecting any other keys.  The
// prefixes also give us some sanity, so that we completely avoid inter-index key conflicts.
// NOTE: These must not conflict with dateRangeToHeightSearchKeyPrefix defined in metanode.
const addressToHeightPrefix = "address:height:"

func fmtAddressToHeight(addr string) string {
	return addressToHeightPrefix + addr
}

const blockHashToHeightPrefix = "block.hash:height:"

func fmtBlockHashToHeight(hash string) string {
	return blockHashToHeightPrefix + strings.ToLower(hash)
}

const heightToTimestampPrefix = "height:timestamp:"

func fmtHeightToTimestamp(height uint64) string {
	return fmt.Sprintf("%s%d", heightToTimestampPrefix, height)
}

const marketPriceKeysetKey = "marketPriceKeys"
const marketPriceKeysetPrefix = "market.price:"

func fmtMarketPriceTimeKey(timestamp math.Timestamp) string {
	return fmt.Sprintf("%sts:%s", marketPriceKeysetPrefix, timestamp.String())
}

func fmtMarketPriceHeightKey(height uint64) string {
	return fmt.Sprintf("%sh:%d", marketPriceKeysetPrefix, height)
}

const sysvarKeyToValuePrefix = "sysvar.key:value:"

func fmtSysvarKeyToValue(key string) string {
	return sysvarKeyToValuePrefix + key
}

const txHashToHeightPrefix = "tx.hash:height:"

func fmtTxHashToHeight(hash string) string {
	return txHashToHeightPrefix + hash
}

const txTypeToHeightPrefix = "tx.type:height:"

func fmtTxTypeToHeight(typeName string) string {
	return txTypeToHeightPrefix + strings.ToLower(typeName)
}

const unionPrefix = "union:"

func fmtUnion() string {
	// Use time now to effectively guarantee uniqueness for every caller.
	return fmt.Sprintf("%s%d", unionPrefix, time.Now().UnixNano())
}
