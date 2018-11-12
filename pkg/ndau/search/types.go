package search

// Types common to indexing and searching.

import (
	"fmt"
	"strconv"
	"strings"
)

// TxValueData is used for storing the block height and transaction offset within the block.
type TxValueData struct {
	BlockHeight uint64
	TxOffset uint64
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

	offset, err := strconv.ParseUint(searchValue[separator+1:], 10, 64)
	if err != nil {
		return err
	}

	valueData.BlockHeight = height
	valueData.TxOffset = offset

	return nil
}
