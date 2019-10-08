package backing

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"errors"

	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

//go:generate msgp

// generate noms marshaler implementations for appropriate types
//nomsify Lock

// Lock keeps track of an account's Lock information
type Lock struct {
	NoticePeriod math.Duration `msg:"notice" chain:"91,Lock_NoticePeriod" json:"noticePeriod"`
	// if a lock has not been notified, this is nil
	UnlocksOn *math.Timestamp `msg:"unlock" chain:"92,Lock_UnlocksOn" json:"unlocksOn"`
	Bonus     eai.Rate        `msg:"bonus" chain:"93,Lock_Bonus" json:"bonus"`
}

// NewLock constructs a new lock with appropriate bonus rate looked up
func NewLock(period math.Duration, bonusTable eai.RateTable) *Lock {
	return &Lock{
		NoticePeriod: period,
		Bonus:        bonusTable.RateAt(period),
	}
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

// GetBonusRate implements eai.Lock
func (l *Lock) GetBonusRate() eai.Rate {
	if l != nil {
		return l.Bonus
	}
	return eai.Rate(0)
}

var _ eai.Lock = (*Lock)(nil)

// Notify updates this lock with notification of intent to unlock
func (l *Lock) Notify(blockTime math.Timestamp, weightedAverageAge math.Duration) error {
	if l.UnlocksOn != nil {
		return errors.New("already notified")
	}
	uo := blockTime.Add(l.NoticePeriod)
	l.UnlocksOn = &uo
	return nil
}
