package backing

import (
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	util "github.com/oneiro-ndev/noms-util"
)

// generate msgp interface implementations for AccountData and supporting structs
// we can't generate the streaming interfaces, unfortunately, because the
// signature.* types don't implement those
//go:generate msgp -io=0

// Stake keeps track of an account's staking information
type Stake struct {
	Point   math.Timestamp  `chain:"101,Stake_Point"`
	Address address.Address `chain:"102,Stake_Address"`
}

var _ marshal.Marshaler = (*Stake)(nil)
var _ marshal.Unmarshaler = (*Stake)(nil)

// MarshalNoms implements Marshaler for Stake
func (s Stake) MarshalNoms(vrw nt.ValueReadWriter) (val nt.Value, err error) {
	return marshal.Marshal(vrw, s.toNomsStake())
}

// UnmarshalNoms implements Unmarshaler for Stake
func (s *Stake) UnmarshalNoms(v nt.Value) error {
	n := nomsStake{}
	err := marshal.Unmarshal(v, &n)
	if err != nil {
		return err
	}
	return s.fromNomsStake(n)
}

type nomsStake struct {
	Point   util.Int
	Address string
}

func (s Stake) toNomsStake() nomsStake {
	return nomsStake{
		Point:   util.Int(s.Point),
		Address: s.Address.String(),
	}
}

func (s *Stake) fromNomsStake(n nomsStake) (err error) {
	s.Point = math.Timestamp(n.Point)
	if len(n.Address) > 0 {
		s.Address, err = address.Validate(n.Address)
	}
	return
}

// Settlement tracks a single inbound transaction not yet settled
type Settlement struct {
	Qty math.Ndau `chain:"81,Settlement_Quantity"`
	// Expiry is when these funds are available to be sent
	Expiry math.Timestamp `chain:"82,Settlement_Expiry"`
}

var _ marshal.Marshaler = (*Settlement)(nil)
var _ marshal.Unmarshaler = (*Settlement)(nil)

// MarshalNoms implements Marshaler for Settlement
func (e Settlement) MarshalNoms(vrw nt.ValueReadWriter) (val nt.Value, err error) {
	return marshal.Marshal(vrw, e.toNomsSettlement())
}

// UnmarshalNoms implements Unmarshaler for Settlement
func (e *Settlement) UnmarshalNoms(v nt.Value) error {
	n := nomsSettlement{}
	err := marshal.Unmarshal(v, &n)
	if err != nil {
		return err
	}
	e.fromNomsSettlement(n)
	return nil
}

type nomsSettlement struct {
	Qty    util.Int
	Expiry util.Int
}

func (e Settlement) toNomsSettlement() nomsSettlement {
	return nomsSettlement{
		Qty:    util.Int(e.Qty),
		Expiry: util.Int(e.Expiry),
	}
}

func (e *Settlement) fromNomsSettlement(n nomsSettlement) {
	e.Qty = math.Ndau(n.Qty)
	e.Expiry = math.Timestamp(n.Expiry)
}

// SettlementSettings tracks the settlement settings for outbound transactions
type SettlementSettings struct {
	Period    math.Duration   `chain:"111,SettlementSettings_Period"`
	ChangesAt *math.Timestamp `chain:"112,SettlementSettings_ChangesAt"`
	Next      *math.Duration  `chain:"113,SettlementSettings_Next"`
}

var _ marshal.Marshaler = (*SettlementSettings)(nil)
var _ marshal.Unmarshaler = (*SettlementSettings)(nil)

// MarshalNoms implements Marshaler for SettlementSettings
func (e SettlementSettings) MarshalNoms(vrw nt.ValueReadWriter) (val nt.Value, err error) {
	return marshal.Marshal(vrw, e.toNomsSettlementSettings())
}

// UnmarshalNoms implements Unmarshaler for SettlementSettings
func (e *SettlementSettings) UnmarshalNoms(v nt.Value) error {
	n := nomsSettlementSettings{}
	err := marshal.Unmarshal(v, &n)
	if err != nil {
		return err
	}
	e.fromNomsSettlementSettings(n)
	return nil
}

type nomsSettlementSettings struct {
	Duration  util.Int
	HasUpdate bool
	ChangesAt util.Int
	Next      util.Int
}

func (e SettlementSettings) toNomsSettlementSettings() nomsSettlementSettings {
	nes := nomsSettlementSettings{
		Duration:  util.Int(e.Period),
		HasUpdate: e.ChangesAt != nil && e.Next != nil,
	}
	if nes.HasUpdate {
		nes.ChangesAt = util.Int(*e.ChangesAt)
		nes.Next = util.Int(*e.Next)
	}
	return nes
}

func (e *SettlementSettings) fromNomsSettlementSettings(n nomsSettlementSettings) {
	e.Period = math.Duration(n.Duration)
	if n.HasUpdate {
		ts := math.Timestamp(n.ChangesAt)
		e.ChangesAt = &ts
		n := math.Duration(n.Next)
		e.Next = &n
	} else {
		e.ChangesAt = nil
		e.Next = nil
	}
}

// NewAccountData creates a new AccountData struct
//
// The zero value of AccountData is not useful, because AccountData needs
// to have non-zero values for LastEAIUpdate and LastWAAUpdate if its EAI
// and WAA calculations are to be accurate.
//
// Unfortunately, go being go, we can't require that this method is used,
// but we can provide it to make it easier to do the right thing.
func NewAccountData(blockTime math.Timestamp) AccountData {
	return AccountData{
		LastEAIUpdate: blockTime,
		LastWAAUpdate: blockTime,
	}
}

// AccountData contains all the information the node needs to take action on a particular account.
//
// See the whitepaper: https://github.com/oneiro-ndev/whitepapers/blob/master/node_incentives/transactions.md#wallet-data
type AccountData struct {
	Balance             math.Ndau             `chain:"61,Acct_Balance"`
	TransferKeys        []signature.PublicKey `chain:"62,Acct_TransferKeys"`
	RewardsTarget       *address.Address      `chain:"63,Acct_RewardsTarget"`
	IncomingRewardsFrom []address.Address     `chain:"64,Acct_IncomingRewardsFrom"`
	DelegationNode      *address.Address      `chain:"65,Acct_DelegationNode"`
	Lock                *Lock                 `chain:"."`
	Stake               *Stake                `chain:"."`
	LastEAIUpdate       math.Timestamp        `chain:"66,Acct_LastEAIUpdate"`
	LastWAAUpdate       math.Timestamp        `chain:"67,Acct_LastWAAUpdate"`
	WeightedAverageAge  math.Duration         `chain:"68,Acct_WeightedAverageAge"`
	Sequence            uint64
	Settlements         []Settlement       `chain:"."`
	SettlementSettings  SettlementSettings `chain:"."`
}

var _ marshal.Marshaler = (*AccountData)(nil)
var _ marshal.Unmarshaler = (*AccountData)(nil)

// MarshalNoms implements Marshaler for AccountData
func (ad AccountData) MarshalNoms(vrw nt.ValueReadWriter) (val nt.Value, err error) {
	nad, err := ad.toNomsAccountData(vrw)
	if err != nil {
		return nil, err
	}
	return marshal.Marshal(vrw, nad)
}

// UnmarshalNoms implements Unmarshaler for AccountData
func (ad *AccountData) UnmarshalNoms(v nt.Value) error {
	n := nomsAccountData{}
	err := marshal.Unmarshal(v, &n)
	if err != nil {
		return err
	}
	return ad.fromNomsAccountData(n)
}

type nomsAccountData struct {
	Balance             util.Int
	TransferKeys        []nt.Blob
	HasRewardsTarget    bool
	RewardsTarget       nt.String
	IncomingRewardsFrom []nt.String
	HasDelegationNode   bool
	DelegationNode      nt.String
	HasLock             bool
	Lock                Lock
	HasStake            bool
	Stake               Stake
	LastEAIUpdate       util.Int
	LastWAAUpdate       util.Int
	WeightedAverageAge  util.Int
	Sequence            util.Int
	Settlements         []Settlement
	SettlementSettings  SettlementSettings
}

func (ad AccountData) toNomsAccountData(vrw nt.ValueReadWriter) (nomsAccountData, error) {
	nad := nomsAccountData{
		Balance:            util.Int(ad.Balance),
		HasRewardsTarget:   ad.RewardsTarget != nil,
		HasDelegationNode:  ad.DelegationNode != nil,
		HasLock:            ad.Lock != nil,
		HasStake:           ad.Stake != nil,
		LastEAIUpdate:      util.Int(ad.LastEAIUpdate),
		LastWAAUpdate:      util.Int(ad.LastWAAUpdate),
		WeightedAverageAge: util.Int(ad.WeightedAverageAge),
		Sequence:           util.Int(ad.Sequence),
		Settlements:        ad.Settlements,
		SettlementSettings: ad.SettlementSettings,
	}
	for _, tk := range ad.TransferKeys {
		tkBytes, err := tk.Marshal()
		if err != nil {
			return nomsAccountData{}, err
		}
		nad.TransferKeys = append(nad.TransferKeys, util.Blob(vrw, tkBytes))
	}
	if nad.HasRewardsTarget {
		nad.RewardsTarget = nt.String(ad.RewardsTarget.String())
	}
	for _, irf := range ad.IncomingRewardsFrom {
		nad.IncomingRewardsFrom = append(nad.IncomingRewardsFrom, nt.String(irf.String()))
	}
	if nad.HasDelegationNode {
		nad.DelegationNode = nt.String(ad.DelegationNode.String())
	}
	if nad.HasLock {
		nad.Lock = *ad.Lock
	}
	if nad.HasStake {
		nad.Stake = *ad.Stake
	}
	return nad, nil
}

func (ad *AccountData) fromNomsAccountData(n nomsAccountData) (err error) {
	ad.Balance = math.Ndau(n.Balance)
	if err != nil {
		*ad = AccountData{}
		return err
	}
	for _, ntk := range n.TransferKeys {
		tkBytes, err := util.Unblob(ntk)
		if err != nil {
			*ad = AccountData{}
			return err
		}
		tk := signature.PublicKey{}
		err = tk.Unmarshal(tkBytes)
		if err != nil {
			*ad = AccountData{}
			return err
		}
		ad.TransferKeys = append(ad.TransferKeys, tk)
	}
	if n.HasRewardsTarget {
		ad.RewardsTarget = new(address.Address)
		*ad.RewardsTarget, err = address.Validate(string(n.RewardsTarget))
		if err != nil {
			*ad = AccountData{}
			return err
		}
	}
	for _, irf := range n.IncomingRewardsFrom {
		addr, err := address.Validate(string(irf))
		if err != nil {
			*ad = AccountData{}
			return errors.Wrap(err, "invalid incoming rewards from: "+string(irf))
		}
		ad.IncomingRewardsFrom = append(ad.IncomingRewardsFrom, addr)
	}
	if n.HasDelegationNode {
		ad.DelegationNode = new(address.Address)
		*ad.DelegationNode, err = address.Validate(string(n.DelegationNode))
		if err != nil {
			*ad = AccountData{}
			return err
		}
	}
	if n.HasLock {
		ad.Lock = &n.Lock
	} else {
		ad.Lock = nil
	}
	if n.HasStake {
		ad.Stake = &n.Stake
	} else {
		ad.Stake = nil
	}
	ad.LastEAIUpdate = math.Timestamp(n.LastEAIUpdate)
	ad.LastWAAUpdate = math.Timestamp(n.LastWAAUpdate)
	ad.WeightedAverageAge = math.Duration(n.WeightedAverageAge)
	ad.Sequence = uint64(n.Sequence)
	ad.Settlements = n.Settlements
	ad.SettlementSettings = n.SettlementSettings
	return nil
}
