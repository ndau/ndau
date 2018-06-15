package ndau

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
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
	Source      address.Address
	Destination address.Address
	Qty         math.Ndau
	Sequence    uint64
	Signature   []byte
}

// static assert that GTValidatorChange is metatx.Transactable
var _ metatx.Transactable = (*Transfer)(nil)

// SigningKeyKind is the kind of key used to sign a ChangeTransferKey
type SigningKeyKind byte

const (
	// SigningKeyOwnership indicates that the ownership key is used to sign the ChangeTransferKey transaction
	SigningKeyOwnership SigningKeyKind = 0x01
	// SigningKeyTransfer indicates that the previous transfer key is used to sign the ChangeTransferKey transaction
	SigningKeyTransfer SigningKeyKind = 0x02
)

// A ChangeTransferKey transaction is used to set a transfer key
//
// It may be signed with the account ownership key or the previous public key.
// KeyKind is used to identify which of these are in use.
type ChangeTransferKey struct {
	Target     address.Address
	NewKey     []byte
	SigningKey []byte
	KeyKind    SigningKeyKind
	Signature  []byte
}

var _ metatx.Transactable = (*ChangeTransferKey)(nil)
