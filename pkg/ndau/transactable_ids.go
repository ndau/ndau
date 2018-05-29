package ndau

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
)

// TxIDs is a map which defines canonical numeric ids for each transactable type.
var TxIDs = map[metatx.TxID]metatx.Transactable{
	metatx.TxID(0xff): &GTValidatorChange{},
}
