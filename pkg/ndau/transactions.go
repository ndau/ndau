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
	metatx.TxID(2):    &ChangeValidation{},
	metatx.TxID(3):    &ReleaseFromEndowment{},
	metatx.TxID(4):    &ChangeSettlementPeriod{},
	metatx.TxID(5):    &Delegate{},
	metatx.TxID(6):    &CreditEAI{},
	metatx.TxID(7):    &Lock{},
	metatx.TxID(8):    &Notify{},
	metatx.TxID(9):    &SetRewardsDestination{},
	metatx.TxID(10):   &ClaimAccount{},
	metatx.TxID(11):   &Stake{},
	metatx.TxID(12):   &RegisterNode{},
	metatx.TxID(13):   &NominateNodeReward{},
	metatx.TxID(14):   &ClaimNodeReward{},
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

// static assert that GTValidatorChange is ndauTransactable
var _ metatx.Transactable = (*GTValidatorChange)(nil)

// A Transfer is the fundamental transaction of the Ndau chain.
type Transfer struct {
	Source      address.Address       `msg:"src" chain:"1,Tx_Source"`
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination"`
	Qty         math.Ndau             `msg:"qty" chain:"11,Tx_Quantity"`
	Sequence    uint64                `msg:"seq"`
	Signatures  []signature.Signature `msg:"sig"`
}

// static assert that GTValidatorChange is ndauTransactable
var _ ndauTransactable = (*Transfer)(nil)

// A ChangeValidation transaction is used to set transfer keys
type ChangeValidation struct {
	Target           address.Address       `msg:"tgt" chain:"3,Tx_Target"`
	NewKeys          []signature.PublicKey `msg:"key" chain:"31,Tx_NewKeys"`
	ValidationScript []byte                `msg:"val" chain:"32,Tx_ValidationScript"`
	Sequence         uint64                `msg:"seq"`
	Signatures       []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*ChangeValidation)(nil)

// A ReleaseFromEndowment transaction is used to release funds from the
// endowment into an individual account.
//
// It must be signed with the private key corresponding to one of the public
// keys listed in the system variable `ReleaseFromEndowmentKeys`.
type ReleaseFromEndowment struct {
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination"`
	Qty         math.Ndau             `msg:"qty" chain:"11,Tx_Quantity"`
	Sequence    uint64                `msg:"seq"`
	Signatures  []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*ReleaseFromEndowment)(nil)

// A ChangeSettlementPeriod transaction is used to change the settlement period for
// transactions outbound from an account.
type ChangeSettlementPeriod struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target"`
	Period     math.Duration         `msg:"per" chain:"21,Tx_Period"`
	Sequence   uint64                `msg:"seq"`
	Signatures []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*ChangeSettlementPeriod)(nil)

// A Delegate transaction is used to delegate the node which should
// compute EAI for the specified account.
//
// The sequence number must be higher than that of the target Account
type Delegate struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target"`
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node"`
	Sequence   uint64                `msg:"seq"`
	Signatures []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*Delegate)(nil)

// A CreditEAI transaction is used to award EAI.
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
type CreditEAI struct {
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node"`
	Sequence   uint64                `msg:"seq"`
	Signatures []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*CreditEAI)(nil)

// Lock transactions lock the specfied account.
//
// Locked accounts may still receive ndau but may not be the source for transfers.
type Lock struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target"`
	Period     math.Duration         `msg:"per" chain:"21,Tx_Period"`
	Sequence   uint64                `msg:"seq"`
	Signatures []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*Lock)(nil)

// Notify transactions notify that the specified account should be unlocked once
// its notice period expires.
//
// Notified accounts may not receive ndau.
type Notify struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target"`
	Sequence   uint64                `msg:"seq"`
	Signatures []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*Notify)(nil)

// SetRewardsDestination transactions update the rewards target for the specified account.
//
// When the rewards target is empty, EAI and other rewards are deposited to the
// origin account. Otherwise, they are deposited to the specified destination.
type SetRewardsDestination struct {
	Source      address.Address       `msg:"src" chain:"1,Tx_Source"`
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination"`
	Sequence    uint64                `msg:"seq"`
	Signatures  []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*SetRewardsDestination)(nil)

// A ClaimAccount transaction is used to set the initial transfer keys for an account.
//
// It is the only type of transaction which may be signed with the ownership key.
type ClaimAccount struct {
	Target           address.Address       `msg:"tgt"`
	Ownership        signature.PublicKey   `msg:"own"`
	TransferKeys     []signature.PublicKey `msg:"key"`
	ValidationScript []byte                `msg:"val"`
	Sequence         uint64                `msg:"seq"`
	Signature        signature.Signature   `msg:"sig"`
}

var _ ndauTransactable = (*ClaimAccount)(nil)

// A Stake transaction stakes to a node
type Stake struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target"`
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node"`
	Sequence   uint64                `msg:"seq"`
	Signatures []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*Stake)(nil)

// A RegisterNode transaction activates a node
type RegisterNode struct {
	Node               address.Address       `msg:"nod" chain:"4,Tx_Node"`
	DistributionScript []byte                `msg:"dis" chain:"33,Tx_DistributionScript"`
	RPCAddress         string                `msg:"rpc" chain:"34,Tx_RPCAddress"`
	Sequence           uint64                `msg:"seq"`
	Signatures         []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*RegisterNode)(nil)

// A NominateNodeReward transaction signals that a node is probably about to be
// rewarded.
//
// The signatures are checked against an account specified by the
// NominateNodeRewardAddress system variable. That account also specifes
// the validation script, and pays the transaction fee.
type NominateNodeReward struct {
	Random     uint64                `msg:"rnd" chain:"41,Tx_Random"`
	Sequence   uint64                `msg:"seq"`
	Signatures []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*NominateNodeReward)(nil)

// A ClaimNodeReward transaction signals that the named node has been watching
// the blockchain, noticed that it won the nomination, and is up and ready to
// claim its reward.
type ClaimNodeReward struct {
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node"`
	Sequence   uint64                `msg:"seq"`
	Signatures []signature.Signature `msg:"sig"`
}

var _ ndauTransactable = (*ClaimNodeReward)(nil)
