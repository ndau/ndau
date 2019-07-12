package search

// Types common to indexing and searching.

import (
	"encoding/base64"

	"github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

//go:generate msgp

// HeightByBlockHashCommand is a QueryParams command for searching block height by block hash.
const HeightByBlockHashCommand = "heightbyblockhash"

// HeightByTxHashCommand is a QueryParams command for searching block height by tx hash.
const HeightByTxHashCommand = "heightbytxhash"

// QueryParams is a json-friendly struct for passing query terms over endpoints.
type QueryParams struct {
	// App-specific command.
	Command string `json:"command"`

	// A block hash or tx hash (or any other kind of hash), depending on the command.
	Hash string `json:"hash"`
}

// SysvarHistoryParams is a json-friendly struct for the /sysvar/history endpoint.
type SysvarHistoryParams struct {
	Name        string `json:"name"`
	AfterHeight uint64 `json:"afterheight"`
	Limit       int    `json:"limit"`
}

// ValueData is used for skipping duplicate key value pairs while iterating the blockchain.
type ValueData struct {
	height      uint64 `msg:"h"`
	valueBase64 string `msg:"v"`
}

// AccountHistoryParams is a json-friendly struct for the /account/history endpoint.
type AccountHistoryParams struct {
	Address     string `json:"addr"`
	AfterHeight uint64 `json:"afterheight"`
	Limit       int    `json:"limit"`
}

// AccountListParams is a json-friendly struct for the /account/list endpoint.
type AccountListParams struct {
	Address string `json:"addr"`
	After   string `json:"after"`
	Limit   int    `json:"limit"`
}

// TxValueData is used for data about a particular transaction
type TxValueData struct {
	BlockHeight uint64 `json:"height" msg:"h"`
	TxOffset    int    `json:"offset" msg:"o"`
	Fee         uint64 `json:"fee" msg:"f"`
	SIB         uint64 `json:"sib" msg:"s"`
}

// AccountTxValueData is like TxValueData that stores account balance at the associated block.
// We could index a Ref target hash, but that would use more space than just storing the balance.
type AccountTxValueData struct {
	BlockHeight uint64     `msg:"h"`
	TxOffset    int        `msg:"o"`
	Balance     types.Ndau `msg:"b"`
}

// AccountHistoryResponse is the return value from the account history endpoint.
type AccountHistoryResponse struct {
	Txs []AccountTxValueData `msg:"t"`
}

// Marshal the value data into a search value string to index it with its search key string.
func (valueData *ValueData) Marshal() string {
	msgp, err := valueData.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(msgp)
}

// Unmarshal the given search value string that was indexed with its search key string.
func (valueData *ValueData) Unmarshal(searchValue string) error {
	bytes, err := base64.StdEncoding.DecodeString(searchValue)
	if err != nil {
		return errors.Wrap(err, "decoding b64")
	}
	_, err = valueData.UnmarshalMsg(bytes)
	return errors.Wrap(err, "decoding msgp")
}

// Marshal the value data into a search value string to index it with its search key string.
func (valueData *TxValueData) Marshal() string {
	m, err := valueData.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(m)
}

// Unmarshal the given search value string that was indexed with its search key string.
func (valueData *TxValueData) Unmarshal(searchValue string) error {
	bytes, err := base64.StdEncoding.DecodeString(searchValue)
	if err != nil {
		return errors.Wrap(err, "decoding b64")
	}
	_, err = valueData.UnmarshalMsg(bytes)
	return errors.Wrap(err, "decoding msgp")
}

// Marshal the value data into a search value string to index it with its search key string.
func (valueData *AccountTxValueData) Marshal() string {
	m, err := valueData.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(m)
}

// Unmarshal the given search value string that was indexed with its search key string.
func (valueData *AccountTxValueData) Unmarshal(searchValue string) error {
	bytes, err := base64.StdEncoding.DecodeString(searchValue)
	if err != nil {
		return errors.Wrap(err, "decoding b64")
	}
	_, err = valueData.UnmarshalMsg(bytes)
	return errors.Wrap(err, "decoding msgp")
}

// Marshal the account history response into something we can pass over RPC.
func (response *AccountHistoryResponse) Marshal() string {
	m, err := response.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(m)
}

// Unmarshal the account history response from something we received over RPC.
func (response *AccountHistoryResponse) Unmarshal(searchValue string) error {
	bytes, err := base64.StdEncoding.DecodeString(searchValue)
	if err != nil {
		return errors.Wrap(err, "decoding b64")
	}
	_, err = response.UnmarshalMsg(bytes)
	return errors.Wrap(err, "decoding msgp")
}
