package ndau

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// We disable tests because every transaction except GTValidatorChange
// uses the Signature type, which blows up the msgp-generated tests because
// it doesn't work with nil values.
//go:generate msgp -io=0 -tests=0

// TxIDs is a map which defines canonical numeric ids for each transactable type.
var TxIDs = map[metatx.TxID]metatx.Transactable{
	metatx.TxID(1):  &Transfer{},
	metatx.TxID(2):  &ChangeValidation{},
	metatx.TxID(3):  &ReleaseFromEndowment{},
	metatx.TxID(4):  &ChangeRecoursePeriod{},
	metatx.TxID(5):  &Delegate{},
	metatx.TxID(6):  &CreditEAI{},
	metatx.TxID(7):  &Lock{},
	metatx.TxID(8):  &Notify{},
	metatx.TxID(9):  &SetRewardsDestination{},
	metatx.TxID(10): &SetValidation{},
	metatx.TxID(11): &Stake{},
	metatx.TxID(12): &RegisterNode{},
	metatx.TxID(13): &NominateNodeReward{},
	metatx.TxID(14): &ClaimNodeReward{},
	metatx.TxID(15): &TransferAndLock{},
	metatx.TxID(16): &CommandValidatorChange{},
	metatx.TxID(18): &UnregisterNode{},
	metatx.TxID(19): &Unstake{},
	metatx.TxID(20): &Issue{},
	metatx.TxID(21): &CreateChildAccount{},
	metatx.TxID(22): &RecordPrice{},
	metatx.TxID(23): &SetSysvar{},
	metatx.TxID(24): &SetStakeRules{},
	metatx.TxID(25): &RecordEndowmentNAV{},
	metatx.TxID(26): &ResolveStake{},
	metatx.TxID(30): &ChangeSchema{},
}

// A Transfer is the fundamental transaction of the Ndau chain.
type Transfer struct {
	Source      address.Address       `msg:"src" chain:"1,Tx_Source" json:"source"`
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination" json:"destination"`
	Qty         math.Ndau             `msg:"qty" chain:"11,Tx_Quantity" json:"qty"`
	Sequence    uint64                `msg:"seq" json:"sequence"`
	Signatures  []signature.Signature `msg:"sig" json:"signatures"`
}

// static assert that Transfer is NTransactable
var _ NTransactable = (*Transfer)(nil)

// A ChangeValidation transaction is used to set validation rules
type ChangeValidation struct {
	Target           address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	NewKeys          []signature.PublicKey `msg:"key" chain:"31,Tx_NewKeys" json:"new_keys"`
	ValidationScript []byte                `msg:"val" chain:"32,Tx_ValidationScript" json:"validation_script"`
	Sequence         uint64                `msg:"seq" json:"sequence"`
	Signatures       []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*ChangeValidation)(nil)

// A ReleaseFromEndowment transaction is used to release funds from the
// endowment into an individual account.
//
// It must be signed with the private key corresponding to one of the public
// keys listed in the system variable `ReleaseFromEndowmentKeys`.
type ReleaseFromEndowment struct {
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination" json:"destination"`
	Qty         math.Ndau             `msg:"qty" chain:"11,Tx_Quantity" json:"qty"`
	Sequence    uint64                `msg:"seq" json:"sequence"`
	Signatures  []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*ReleaseFromEndowment)(nil)

// A ChangeRecoursePeriod transaction is used to change the recourse period for
// transactions outbound from an account.
type ChangeRecoursePeriod struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Period     math.Duration         `msg:"per" chain:"21,Tx_Period" json:"period"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*ChangeRecoursePeriod)(nil)

// A Delegate transaction is used to delegate the node which should
// compute EAI for the specified account.
//
// The sequence number must be higher than that of the target Account
type Delegate struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node" json:"node"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*Delegate)(nil)

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
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node" json:"node"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*CreditEAI)(nil)

// Lock transactions lock the specfied account.
//
// Locked accounts may still receive ndau but may not be the source for transfers.
type Lock struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Period     math.Duration         `msg:"per" chain:"21,Tx_Period" json:"period"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*Lock)(nil)

// Notify transactions notify that the specified account should be unlocked once
// its notice period expires.
//
// Notified accounts may not receive ndau.
type Notify struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*Notify)(nil)

// SetRewardsDestination transactions update the rewards target for the specified account.
//
// When the rewards target is empty, EAI and other rewards are deposited to the
// origin account. Otherwise, they are deposited to the specified destination.
type SetRewardsDestination struct {
	Target      address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination" json:"destination"`
	Sequence    uint64                `msg:"seq" json:"sequence"`
	Signatures  []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*SetRewardsDestination)(nil)

// A SetValidation transaction is used to set the initial validation rules for an account.
//
// It is the only type of transaction which may be signed with the ownership key.
type SetValidation struct {
	Target           address.Address       `msg:"tgt" json:"target"`
	Ownership        signature.PublicKey   `msg:"own" json:"ownership"`
	ValidationKeys   []signature.PublicKey `msg:"key" json:"validation_keys"`
	ValidationScript []byte                `msg:"val" json:"validation_script"`
	Sequence         uint64                `msg:"seq" json:"sequence"`
	Signature        signature.Signature   `msg:"sig" json:"signature"`
}

var _ NTransactable = (*SetValidation)(nil)

// A CreateChildAccount transaction is used to set the initial validation rules for a child account,
// and link the target account as its parent.
type CreateChildAccount struct {
	Target                address.Address       `msg:"tgt"  json:"target"`
	Child                 address.Address       `msg:"chd"  json:"child"`
	ChildOwnership        signature.PublicKey   `msg:"cown" json:"child_ownership"`
	ChildSignature        signature.Signature   `msg:"csig" json:"child_signature"`
	ChildRecoursePeriod   math.Duration         `msg:"cper" json:"child_recourse_period"`
	ChildValidationKeys   []signature.PublicKey `msg:"ckey" json:"child_validation_keys"`
	ChildValidationScript []byte                `msg:"cval" json:"child_validation_script"`
	ChildDelegationNode   address.Address       `msg:"nod"  json:"child_delegation_node"`
	Sequence              uint64                `msg:"seq"  json:"sequence"`
	Signatures            []signature.Signature `msg:"sig"  json:"signatures"`
}

var _ NTransactable = (*CreateChildAccount)(nil)

// TransferAndLock allows a transaction where the received amount is locked
// for a specified period. It can only be sent to accounts that did not
// previously exist on the blockchain.
type TransferAndLock struct {
	Source      address.Address       `msg:"src" chain:"1,Tx_Source" json:"source"`
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination" json:"destination"`
	Qty         math.Ndau             `msg:"qty" chain:"11,Tx_Quantity" json:"qty"`
	Period      math.Duration         `msg:"per" chain:"21,Tx_Period" json:"period"`
	Sequence    uint64                `msg:"seq" json:"sequence"`
	Signatures  []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*TransferAndLock)(nil)

// A Stake transaction stakes to a node
type Stake struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Rules      address.Address       `msg:"rul" chain:"8,Tx_Rules" json:"rules"`
	StakeTo    address.Address       `msg:"sto" chain:"5,Tx_StakeTo" json:"stake_to"`
	Qty        math.Ndau             `msg:"qty" chain:"11,Tx_Quantity" json:"qty"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*Stake)(nil)

// A RegisterNode transaction activates a node
type RegisterNode struct {
	Node               address.Address       `msg:"nod" chain:"4,Tx_Node" json:"node"`
	DistributionScript []byte                `msg:"dis" chain:"33,Tx_DistributionScript" json:"distribution_script"`
	Ownership          signature.PublicKey   `msg:"own" chain:"34,Tx_Ownership" json:"ownership"`
	Sequence           uint64                `msg:"seq" json:"sequence"`
	Signatures         []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*RegisterNode)(nil)

// A NominateNodeReward transaction signals that a node is probably about to be
// rewarded.
//
// The signatures are checked against an account specified by the
// NominateNodeRewardAddress system variable. That account also specifes
// the validation script, and pays the transaction fee.
type NominateNodeReward struct {
	Random     int64                 `msg:"rnd" chain:"41,Tx_Random" json:"random"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*NominateNodeReward)(nil)

// A ClaimNodeReward transaction signals that the named node has been watching
// the blockchain, noticed that it won the nomination, and is up and ready to
// claim its reward.
type ClaimNodeReward struct {
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node" json:"node"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*ClaimNodeReward)(nil)

// A CommandValidatorChange changes a validator's power by fiat.
//
// Like NNR and RFE, there is a special account whose address is stored
// as a system variable, which both authenticates this transaction
// and pays its associated fees.
//
// This transaction replaces GTValidatorChange, which was too insecure
// to ever actually deploy with.
type CommandValidatorChange struct {
	// Node must previously have been registered with RegisterNode
	Node address.Address `msg:"nod" chain:"4,Tx_Node" json:"node"`

	// Power is an arbitrary integer with no intrinsic
	// meaning; during the Global Trust period, it
	// can be literally whatever. Setting it to 0
	// removes the validator.
	Power int64 `msg:"pow" chain:"17,Tx_Power" json:"power"`

	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*CommandValidatorChange)(nil)

// An UnregisterNode transaction deactivates a node
type UnregisterNode struct {
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node" json:"node"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*UnregisterNode)(nil)

// An Unstake transaction stakes to a node
type Unstake struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Rules      address.Address       `msg:"rul" chain:"8,Tx_Rules" json:"rules"`
	StakeTo    address.Address       `msg:"sto" chain:"5,Tx_StakeTo" json:"stake_to"`
	Qty        math.Ndau             `msg:"qty" chain:"11,Tx_Quantity" json:"qty"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*Unstake)(nil)

// An Issue transaction is the second half of the primary sales process.
//
// See https://github.com/oneiro-ndev/ndau/issues/229 for details.
//
// The signatures are checked against an account specified by the
// ReleaseFromEndowmentAddress system variable. That account also specifes
// the validation script, and pays the transaction fee.
type Issue struct {
	Qty        math.Ndau             `msg:"qty" chain:"11,Tx_Quantity" json:"qty"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*Issue)(nil)

// A RecordPrice transaction records the current market price of Ndau.
//
// This data is used to calculate the current SIB in effect.
//
// Its signatures are checked against an account specified by the
// RecordPriceAddress system variable. That account also specifies the validation
// script, and pays the transaction fee.
type RecordPrice struct {
	MarketPrice pricecurve.Nanocent   `msg:"prc" chain:"11,Tx_Quantity" json:"market_price"`
	Sequence    uint64                `msg:"seq" json:"sequence"`
	Signatures  []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*RecordPrice)(nil)

// A SetSysvar transaction sets a system variable.
//
// Its signatures are checked against an account specified by the SetSysvarAddress
// system variable. That account also specifies the validation script, and pays
// the transaction fee.
type SetSysvar struct {
	Name       string                `msg:"nme" chain:"6,Tx_Name" json:"name"`
	Value      []byte                `msg:"vlu" chain:"7,Tx_Value" json:"value"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*SetSysvar)(nil)

// A SetStakeRules transaction is used to set or remove stake rules from an account
type SetStakeRules struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	StakeRules []byte                `msg:"krs" chain:"35,Tx_StakeRules" json:"stake_rules"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*SetStakeRules)(nil)

// A ChangeSchema transaction triggers the ndaunode to shut down.
//
// This is used to enable versioning upgrades which change the noms schema.
type ChangeSchema struct {
	// SchemaVersion is advisory and not checked or retained by the blockchain.
	// It is intended to be read by humans replaying the blockchain.
	SchemaVersion string                `msg:"sav" json:"schema_version"`
	Sequence      uint64                `msg:"seq" json:"sequence"`
	Signatures    []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*ChangeSchema)(nil)

// A RecordEndowmentNAV transaction records the current Net Asset Value of the ndau endowment.
//
// This data is used to calculate the current SIB in effect.
//
// Its signatures are checked against an account specified by the
// RecordEndowmentNAVAddress system variable. That account also specifies the
// validation script, and pays the transaction fee.
type RecordEndowmentNAV struct {
	NAV        pricecurve.Nanocent   `msg:"nav" chain:"11,Tx_Quantity" json:"nav"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*RecordEndowmentNAV)(nil)

// A ResolveStake transaction is submitted by a stake rules account to resolve a stake.
//
// Most of the time this will not be required: stakers will eventually `Unstake`, be
// subject to some delay, and then have their staked balance automatically return to
// spendable status.
//
// ResolveStake is not for those transactions. It is for those times when an account
// or group of accounts must have its stake slashed, and for dispute resolution.
type ResolveStake struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"` // primary staker
	Rules      address.Address       `msg:"rul" chain:"8,Tx_Rules" json:"rules"`
	Burn       uint8                 `msg:"brn" chain:"12,Tx_Burn" json:"burn"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}

var _ NTransactable = (*ResolveStake)(nil)
