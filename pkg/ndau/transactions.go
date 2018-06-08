package ndau

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

//go:generate msgp

// TxIDs is a map which defines canonical numeric ids for each transactable type.
var TxIDs = map[metatx.TxID]metatx.Transactable{
	metatx.TxID(1):    &Transfer{},
	metatx.TxID(0xff): &GTValidatorChange{},
}

// A GTValidatorChange is a Globally Trusted Validator Change.
//
// No attempt is made to validate the validator change;
// nobody watches the watchmen.
//
// THIS IS DANGEROUS AND MUST BE DISABLED PRIOR TO RELEASE
type GTValidatorChange struct {
	// No information about the public key format is
	// currently available.
	PublicKey []byte

	// Power is an arbitrary integer with no intrinsic
	// meaning; during the Global Trust period, it
	// can be literally whatever. Setting it to 0
	// removes the validator.
	Power int64
}

// static assert that GTValidatorChange is metatx.Transactable
var _ metatx.Transactable = (*GTValidatorChange)(nil)

// A Transfer is the fundamental transaction of the Ndau chain.
type Transfer struct {
	Timestamp   math.Timestamp
	Source      string // should be address.Address once that PR goes throguh
	Destination string // should be address.Address once that PR goes throguh
	Qty         math.Ndau
	Sequence    uint64
	Signature   []byte
}

// static assert that GTValidatorChange is metatx.Transactable
var _ metatx.Transactable = (*Transfer)(nil)
