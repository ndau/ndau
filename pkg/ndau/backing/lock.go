package backing

import (
	"errors"

	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	util "github.com/oneiro-ndev/noms-util"
)

//go:generate msgp

// Lock keeps track of an account's Lock information
type Lock struct {
	NoticePeriod math.Duration `msg:"notice" chain:"91,Lock_NoticePeriod"`
	// if a lock has not been notified, this is nil
	UnlocksOn *math.Timestamp `msg:"unlock" chain:"92,Lock_UnlocksOn"`
}

// GetNoticePeriod implements eai.Lock
func (l *Lock) GetNoticePeriod() math.Duration {
	if l != nil {
		return l.NoticePeriod
	}
	return 0
}

// GetUnlocksOn implements eai.Lock
func (l *Lock) GetUnlocksOn() *math.Timestamp {
	if l != nil {
		return l.UnlocksOn
	}
	return nil
}

var _ eai.Lock = (*Lock)(nil)

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
	UnlocksOn  util.Int
}

func (l Lock) toNomsLock() nomsLock {
	nl := nomsLock{
		Duration:   util.Int(l.NoticePeriod),
		IsNotified: l.UnlocksOn != nil,
	}
	if l.UnlocksOn != nil {
		nl.UnlocksOn = util.Int(*l.UnlocksOn)
	}
	return nl
}

func (l *Lock) fromNomsLock(nl nomsLock) {
	l.NoticePeriod = math.Duration(nl.Duration)
	if nl.IsNotified {
		ts := math.Timestamp(nl.UnlocksOn)
		l.UnlocksOn = &ts
	} else {
		l.UnlocksOn = nil
	}
}

// Notify updates this lock with notification of intent to unlock
func (l *Lock) Notify(blockTime math.Timestamp, weightedAverageAge math.Duration) error {
	if l.UnlocksOn != nil {
		return errors.New("already notified")
	}
	uo := blockTime.Add(l.NoticePeriod)
	l.UnlocksOn = &uo
	return nil
}
