package backing

import (
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"

	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	util "github.com/oneiro-ndev/noms-util"
)

// Lock keeps track of an account's Lock information
type Lock struct {
	NoticePeriod math.Duration
	// if a lock has not been notified, this is nil
	NotifiedOn *math.Timestamp
}

var _ marshal.Marshaler = (*Lock)(nil)
var _ marshal.Unmarshaler = (*Lock)(nil)

// MarshalNoms implements Marshaler for lock
func (l Lock) MarshalNoms(vrw nt.ValueReadWriter) (val nt.Value, err error) {
	return marshal.Marshal(vrw, l.toNomsLock())
}

// UnmarshalNoms implements Unmarshaler for lock
func (l *Lock) UnmarshalNoms(v nt.Value) error {
	nl := nomsLock{}
	err := marshal.Unmarshal(v, &nl)
	if err != nil {
		return err
	}
	l.fromNomsLock(nl)
	return nil
}

type nomsLock struct {
	Duration   util.Int
	IsNotified bool
	NotifiedOn util.Int
}

func (l Lock) toNomsLock() nomsLock {
	nl := nomsLock{
		Duration:   util.Int(l.NoticePeriod),
		IsNotified: l.NotifiedOn != nil,
	}
	if l.NotifiedOn != nil {
		nl.NotifiedOn = util.Int(*l.NotifiedOn)
	}
	return nl
}

func (l *Lock) fromNomsLock(nl nomsLock) {
	l.NoticePeriod = math.Duration(nl.Duration)
	if nl.IsNotified {
		ts := math.Timestamp(nl.NotifiedOn)
		l.NotifiedOn = &ts
	} else {
		l.NotifiedOn = nil
	}
}

// Stake keeps track of an account's staking information
type Stake struct {
	Point   math.Timestamp
	Address string
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
	s.fromNomsStake(n)
	return nil
}

type nomsStake struct {
	Point   util.Int
	Address string
}

func (s Stake) toNomsStake() nomsStake {
	return nomsStake{
		Point:   util.Int(s.Point),
		Address: s.Address,
	}
}

func (s *Stake) fromNomsStake(n nomsStake) {
	s.Point = math.Timestamp(n.Point)
	s.Address = n.Address
}

// Escrow tracks a single transaction of incoming escrow
type Escrow struct {
	Qty math.Ndau
	// Expiry is when these funds are available to be sent
	Expiry math.Timestamp
}

var _ marshal.Marshaler = (*Escrow)(nil)
var _ marshal.Unmarshaler = (*Escrow)(nil)

// MarshalNoms implements Marshaler for Escrow
func (e Escrow) MarshalNoms(vrw nt.ValueReadWriter) (val nt.Value, err error) {
	return marshal.Marshal(vrw, e.toNomsEscrow())
}

// UnmarshalNoms implements Unmarshaler for Escrow
func (e *Escrow) UnmarshalNoms(v nt.Value) error {
	n := nomsEscrow{}
	err := marshal.Unmarshal(v, &n)
	if err != nil {
		return err
	}
	e.fromNomsEscrow(n)
	return nil
}

type nomsEscrow struct {
	Qty    util.Int
	Expiry util.Int
}

func (e Escrow) toNomsEscrow() nomsEscrow {
	return nomsEscrow{
		Qty:    util.Int(e.Qty),
		Expiry: util.Int(e.Expiry),
	}
}

func (e *Escrow) fromNomsEscrow(n nomsEscrow) {
	e.Qty = math.Ndau(n.Qty)
	e.Expiry = math.Timestamp(n.Expiry)
}

// EscrowSettings tracks the escrow settings for outbound transactions
type EscrowSettings struct {
	Duration  math.Duration
	ChangesAt *math.Timestamp
	Next      *math.Duration
}

var _ marshal.Marshaler = (*EscrowSettings)(nil)
var _ marshal.Unmarshaler = (*EscrowSettings)(nil)

// MarshalNoms implements Marshaler for EscrowSettings
func (e EscrowSettings) MarshalNoms(vrw nt.ValueReadWriter) (val nt.Value, err error) {
	return marshal.Marshal(vrw, e.toNomsEscrowSettings())
}

// UnmarshalNoms implements Unmarshaler for EscrowSettings
func (e *EscrowSettings) UnmarshalNoms(v nt.Value) error {
	n := nomsEscrowSettings{}
	err := marshal.Unmarshal(v, &n)
	if err != nil {
		return err
	}
	e.fromNomsEscrowSettings(n)
	return nil
}

type nomsEscrowSettings struct {
	Duration  util.Int
	HasUpdate bool
	ChangesAt util.Int
	Next      util.Int
}

func (e EscrowSettings) toNomsEscrowSettings() nomsEscrowSettings {
	nes := nomsEscrowSettings{
		Duration:  util.Int(e.Duration),
		HasUpdate: e.ChangesAt != nil && e.Next != nil,
	}
	if nes.HasUpdate {
		nes.ChangesAt = util.Int(*e.ChangesAt)
		nes.Next = util.Int(*e.Next)
	}
	return nes
}

func (e *EscrowSettings) fromNomsEscrowSettings(n nomsEscrowSettings) {
	e.Duration = math.Duration(n.Duration)
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

// AccountData contains all the information the node needs to take action on a particular account.
//
// See the whitepaper: https://github.com/oneiro-ndev/whitepapers/blob/master/node_incentives/transactions.md#wallet-data
type AccountData struct {
	Balance            math.Ndau
	TransferKey        []byte
	RewardsTarget      string
	DelegationNode     string
	Lock               *Lock
	Stake              *Stake
	LastEAITime        math.Timestamp
	WeightedAverageAge math.Duration
	Sequence           uint64
	Escrows            []Escrow
	EscrowSettings     EscrowSettings
}

var _ marshal.Marshaler = (*AccountData)(nil)
var _ marshal.Unmarshaler = (*AccountData)(nil)

// MarshalNoms implements Marshaler for AccountData
func (ad AccountData) MarshalNoms(vrw nt.ValueReadWriter) (val nt.Value, err error) {
	return marshal.Marshal(vrw, ad.toNomsAccountData(vrw))
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
	Balance            util.Int
	TransferKey        nt.Blob
	RewardsTarget      nt.String
	DelegationNode     nt.String
	HasLock            bool
	Lock               Lock
	HasStake           bool
	Stake              Stake
	LastEAITime        util.Int
	WeightedAverageAge util.Int
	Sequence           util.Int
	Escrows            []Escrow
	EscrowSettings     EscrowSettings
}

func (ad AccountData) toNomsAccountData(vrw nt.ValueReadWriter) nomsAccountData {
	nad := nomsAccountData{
		Balance:            util.Int(ad.Balance),
		TransferKey:        util.Blob(vrw, ad.TransferKey),
		RewardsTarget:      nt.String(ad.RewardsTarget),
		DelegationNode:     nt.String(ad.DelegationNode),
		HasLock:            ad.Lock != nil,
		HasStake:           ad.Stake != nil,
		LastEAITime:        util.Int(ad.LastEAITime),
		WeightedAverageAge: util.Int(ad.WeightedAverageAge),
		Sequence:           util.Int(ad.Sequence),
		Escrows:            ad.Escrows,
		EscrowSettings:     ad.EscrowSettings,
	}
	if nad.HasLock {
		nad.Lock = *ad.Lock
	}
	if nad.HasStake {
		nad.Stake = *ad.Stake
	}
	return nad
}

func (ad *AccountData) fromNomsAccountData(n nomsAccountData) (err error) {
	ad.Balance = math.Ndau(n.Balance)
	ad.TransferKey, err = util.Unblob(n.TransferKey)
	if err != nil {
		*ad = AccountData{}
		return err
	}
	ad.RewardsTarget = string(n.RewardsTarget)
	ad.DelegationNode = string(n.DelegationNode)
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
	ad.LastEAITime = math.Timestamp(n.LastEAITime)
	ad.WeightedAverageAge = math.Duration(n.WeightedAverageAge)
	ad.Sequence = uint64(n.Sequence)
	ad.Escrows = n.Escrows
	ad.EscrowSettings = n.EscrowSettings
	return nil
}
