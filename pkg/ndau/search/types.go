package search

// Types common to indexing and searching.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

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

// AccountHistoryParams is a json-friendly struct for the /account/history endpoint.
type AccountHistoryParams struct {
	Address string `json:"addr"`
	PageIndex int `json:"index"`
	PageSize int `json:"size"`
}

// TxValueData is used for storing the block height and transaction offset within the block.
type TxValueData struct {
	BlockHeight uint64
	TxOffset int
}

// AccountTxValueData is like TxValueData that stores account balance at the associated block.
// We could index a Ref target hash, but that would use more space than just storing the balance.
type AccountTxValueData struct {
	BlockHeight uint64
	TxOffset int
	Balance types.Ndau
}

// AccountHistoryResponse is the return value from the account history endpoint.
type AccountHistoryResponse struct {
	Txs []AccountTxValueData
}

// Marshal the value data into a search value string to index it with its search key string.
func (valueData *TxValueData) Marshal() string {
	return fmt.Sprintf("%d %d", valueData.BlockHeight, valueData.TxOffset)
}

// Unmarshal the given search value string that was indexed with its search key string.
func (valueData *TxValueData) Unmarshal(searchValue string) error {
	separator := strings.Index(searchValue, " ")

	height, err := strconv.ParseUint(searchValue[:separator], 10, 64)
	if err != nil {
		return err
	}

	offset, err := strconv.ParseInt(searchValue[separator+1:], 10, 32)
	if err != nil {
		return err
	}

	valueData.BlockHeight = height
	valueData.TxOffset = int(offset)

	return nil
}

// Marshal the value data into a search value string to index it with its search key string.
func (valueData *AccountTxValueData) Marshal() string {
	return fmt.Sprintf("%d %d %d", valueData.BlockHeight, valueData.TxOffset, valueData.Balance)
}

// Unmarshal the given search value string that was indexed with its search key string.
func (valueData *AccountTxValueData) Unmarshal(searchValue string) error {
	separator1 := strings.Index(searchValue, " ")
	separator2 := strings.LastIndex(searchValue, " ")

	height, err := strconv.ParseUint(searchValue[:separator1], 10, 64)
	if err != nil {
		return err
	}

	offset, err := strconv.ParseInt(searchValue[separator1+1:separator2], 10, 32)
	if err != nil {
		return err
	}

	balance, err := strconv.ParseUint(searchValue[separator2+1:], 10, 64)
	if err != nil {
		return err
	}

	valueData.BlockHeight = height
	valueData.TxOffset = int(offset)
	valueData.Balance = types.Ndau(balance)

	return nil
}

// Marshal the account history response into something we can pass over RPC.
func (response *AccountHistoryResponse) Marshal() string {
	var sb strings.Builder
	for _, valueData := range response.Txs {
		sb.WriteString(valueData.Marshal())
		sb.WriteString(":")
	}
	return sb.String()
}

// Unmarshal the account history response from something we received over RPC.
func (response *AccountHistoryResponse) Unmarshal(searchValue string) error {
	response.Txs = nil
	
	items := strings.Split(searchValue, ":")
	for _, item := range items {
		if item != "" {
			valueData := AccountTxValueData{}
			valueData.Unmarshal(item)
			response.Txs = append(response.Txs, valueData)
		}
	}

	return nil
}
