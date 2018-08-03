package ndau

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

// We disable tests because every transaction except GTValidatorChange
// uses the Signature type, which blows up the msgp-generated tests because
// it doesn't work with nil values.
//go:generate msgp -io=0 -tests=0

// TxIDs is a map which defines canonical numeric ids for each transactable type.
var TxIDs = map[metatx.TxID]metatx.Transactable{
	metatx.TxID(1):    &Transfer{},
	metatx.TxID(2):    &ChangeTransferKey{},
	metatx.TxID(3):    &ReleaseFromEndowment{},
	metatx.TxID(4):    &ChangeSettlementPeriod{},
	metatx.TxID(5):    &Delegate{},
	metatx.TxID(6):    &ComputeEAI{},
	metatx.TxID(7):    &Lock{},
	metatx.TxID(8):    &Notify{},
	metatx.TxID(9):    &SetRewardsTarget{},
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
	Signature   signature.Signature
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
	NewKey     signature.PublicKey
	SigningKey signature.PublicKey
	KeyKind    SigningKeyKind
	Sequence   uint64
	Signature  signature.Signature
}

var _ metatx.Transactable = (*ChangeTransferKey)(nil)

// A ReleaseFromEndowment transaction is used to release funds from the
// endowment into an individual account.
//
// It must be signed with the private key corresponding to one of the public
// keys listed in the system variable `ReleaseFromEndowmentKeys`.
type ReleaseFromEndowment struct {
	Destination address.Address
	Qty         math.Ndau
	TxFeeAcct   address.Address
	Sequence    uint64
	Signature   signature.Signature
}

var _ metatx.Transactable = (*ReleaseFromEndowment)(nil)

// A ChangeSettlementPeriod transaction is used to change the settlement period for
// transactions outbound from an account.
type ChangeSettlementPeriod struct {
	Target    address.Address
	Period    math.Duration
	Sequence  uint64
	Signature signature.Signature
}

var _ metatx.Transactable = (*ChangeSettlementPeriod)(nil)

// A Delegate transaction is used to delegate the node which should
// compute EAI for the specified account.
//
// The sequence number must be higher than that of the target Account
type Delegate struct {
	Account   address.Address
	Delegate  address.Address
	Sequence  uint64
	Signature signature.Signature
}

var _ metatx.Transactable = (*Delegate)(nil)

// A ComputeEAI transaction is used to award EAI.
//
// This transaction is sent electively by any node which has accounts delegated
// to it. It is expected that nodes will arrange to create this transaction on
// a regular schedule.
//
// The transaction doesn't include the actual EAI computations. There are two
// reasons for this:
//   1. All nodes must perform the calculations anyway in order to verify that
//      the transaction is valid. If you're doing the calculations anway, there's
//      not much point in adding them to the transaction in the first place.
//   2. The originating node can't know ahead of time what the official block
//      time will be.
type ComputeEAI struct {
	Node      address.Address
	Sequence  uint64
	Signature signature.Signature
}

var _ metatx.Transactable = (*ComputeEAI)(nil)

// Lock transactions lock the specfied account.
//
// Locked accounts may still receive ndau but may not be the source for transfers.
type Lock struct {
	Account   address.Address
	Period    math.Duration
	Sequence  uint64
	Signature signature.Signature
}

var _ metatx.Transactable = (*Lock)(nil)

// Notify transactions notify that the specified account should be unlocked once
// its notice period expires.
//
// Notified accounts may not receive ndau.
type Notify struct {
	Account   address.Address
	Sequence  uint64
	Signature signature.Signature
}

var _ metatx.Transactable = (*Notify)(nil)

// SetRewardsTarget transactions update the rewards target for the specified account.
//
// When the rewards target is empty, EAI and other rewards are deposited to the
// origin account. Otherwise, they are deposited to the specified destination.
type SetRewardsTarget struct {
	Account     address.Address
	Destination address.Address
	Sequence    uint64
	Signature   signature.Signature
}

var _ metatx.Transactable = (*SetRewardsTarget)(nil)
