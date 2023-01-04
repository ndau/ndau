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
	nt "github.com/ndau/noms/go/types"
	meta "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/pricecurve"
	math "github.com/ndau/ndaumath/pkg/types"
)

// State is primarily a set of accounts
//nomsify State
type State struct {
	// managedVars is map that allows us to hide new fields from noms until they're first set.
	// All new variables must start with "managedVar"; nomsify will generate Get/Set accessors.
	managedVars map[string]struct{}
	// Accounts is a map of all accounts that exist on the blockchain.
	Accounts map[string]AccountData
	// Delegates is a map of strings to a set of strings
	// All strings are addresses
	// the keys are the delegated nodes
	// the values are the addresses of the accounts which those nodes must
	// compute
	Delegates map[string]map[string]struct{}
	// Nodes keeps track of the validator and verifier nodes.
	// The key is the node address. The value is a Node struct.
	Nodes map[string]Node
	// The last node reward nomination is necessary in state because this
	// governs the validity of upcoming node reward nominations; there's
	// a minimum interval between them.
	LastNodeRewardNomination math.Timestamp
	// Node rewards are a bit complex. They're accumulated with every
	// CreditEAI transaction into the PendingNodeReward variable. On
	// NominateNodeReward, the balance in PendingNodeReward is moved into
	// UnclaimedNodeReward, because there may be further CreditEAI transactions,
	// which have to be stored up for the subsequent node.
	// ClaimNodeReward transactions actually claim the unclaimed node reward;
	// otherwise, it's overwritten at the next Nominate tx.
	PendingNodeReward   math.Ndau
	UnclaimedNodeReward math.Ndau
	// Of course, we have to keep track of which node has acutally won
	NodeRewardWinner *address.Address
	// TotalRFE is the sum of all RFE transactions.
	// It is also updated by the genesis program: it's initialized with the
	// implied RFEs for genesis accounts
	TotalRFE math.Ndau
	// TotalIssue is the sum of all Issue transactions.
	TotalIssue math.Ndau
	// SIB is the current burn rate applied to all transfers.
	SIB eai.Rate
	// TotalBurned is the sum of all burned ndau
	TotalBurned math.Ndau
	// These prices are preserved here just to assist downstream consumers
	// have more trust in the SIB calculations.
	MarketPrice            pricecurve.Nanocent
	TargetPrice            pricecurve.Nanocent
	managedVarEndowmentNAV pricecurve.Nanocent
	// System variables are all stored here. A system variable is a named
	// msgp-encoded object. It is safe to assume that all keys are valid utf-8.
	Sysvars map[string][]byte
}

// make sure State is a metaapp.State
var _ meta.State = (*State)(nil)

// Init satisfies meta.State
func (s *State) Init(nt.ValueReadWriter) {
	s.Accounts = make(map[string]AccountData)
	s.Delegates = make(map[string]map[string]struct{})
	s.Nodes = make(map[string]Node)
}

// GetAccount returns the account at the requested address.
//
// If the account does not already exist, a fresh one is created.
//
// This function is necessary because account zero values are not valid:
// the `Last*Update` fields must be initialized with the current block time.
//
// The boolean return value is true when the account previously existed;
// false when it is new.
func (s *State) GetAccount(
	address address.Address,
	blockTime math.Timestamp,
	defaultRecoursePeriod math.Duration,
) (AccountData, bool) {
	data, hasAccount := s.Accounts[address.String()]
	if !hasAccount {
		data = NewAccountData(blockTime, defaultRecoursePeriod)
	}
	return data, hasAccount
}

// PayReward updates the state of the target address to add the given qty of ndau, following
// the link to any specified rewards target. If the rewards target account does not previously exist,
// it will be created. Returns a list of accounts whose state was updated in some way.
// Note that if the reward is redirected, it is still the original account whose lastEAIUpdate time
// is changed.
// If isEAI we change WAA only if it's not the same account (per the rules of EAI).
// If isEAI and resetUncreditedEAI we set the source account's UncreditedEAI to 0.
// If it's redirected we change WAA for the target account and do nothing to the source.
// If it's not redirected we change WAA only if it's not EAI.
// After updating the balance in an account, this also updates currency seat
// information for that account.
func (s *State) PayReward(
	srcAddress address.Address,
	reward math.Ndau,
	blockTime math.Timestamp,
	defaultRecoursePeriod math.Duration,
	isEAI bool,
	resetUncreditedEAI bool,
) ([]address.Address, error) {
	var err error
	srcAccount, _ := s.GetAccount(srcAddress, blockTime, defaultRecoursePeriod)

	if isEAI && resetUncreditedEAI {
		// the uncredited EAI is being accounted for, so reset it
		srcAccount.UncreditedEAI = 0
	}

	if srcAccount.RewardsTarget != nil {
		// rewards are being redirected, so get the target account
		tgtAddress := *(srcAccount.RewardsTarget)
		tgtAccount, _ := s.GetAccount(tgtAddress, blockTime, defaultRecoursePeriod)
		// recalc WAA
		err = tgtAccount.WeightedAverageAge.UpdateWeightedAverageAge(
			blockTime.Since(tgtAccount.LastWAAUpdate),
			reward,
			tgtAccount.Balance,
		)
		if err != nil {
			return nil, err
		}
		tgtAccount.LastWAAUpdate = blockTime
		// now we can update the balance
		tgtAccount.Balance, err = tgtAccount.Balance.Add(math.Ndau(reward))
		if err != nil {
			return nil, err
		}
		tgtAccount.UpdateCurrencySeat(blockTime)
		// and store it back into the state
		s.Accounts[tgtAddress.String()] = tgtAccount
		// iff this was EAI, update the source account LastEAIUpdate
		if isEAI {
			srcAccount.LastEAIUpdate = blockTime
			s.Accounts[srcAddress.String()] = srcAccount
			// return both accounts that were changed
			return []address.Address{srcAddress, tgtAddress}, nil
		}
		// it wasn't EAI, but it *was* redirected, so we don't have to update the source
		// at all; just return the target's address.
		return []address.Address{tgtAddress}, nil
	}

	// if we get here, we're not redirected and can do everything to the source account
	if !isEAI {
		// it's not EAI so update WAA but not EAI time
		err = srcAccount.WeightedAverageAge.UpdateWeightedAverageAge(
			blockTime.Since(srcAccount.LastWAAUpdate),
			reward,
			srcAccount.Balance,
		)
		if err != nil {
			return nil, err
		}
		srcAccount.LastWAAUpdate = blockTime
	} else {
		// it IS EAI so update EAI time but not WAA
		srcAccount.LastEAIUpdate = blockTime
	}
	srcAccount.Balance, err = srcAccount.Balance.Add(math.Ndau(reward))
	if err != nil {
		return nil, err
	}
	srcAccount.UpdateCurrencySeat(blockTime)
	// the only account we modify is src
	s.Accounts[srcAddress.String()] = srcAccount
	return []address.Address{srcAddress}, nil
}

// TotalDelegationTo computes the total amount of ndau delegated to a particular node
func (s *State) TotalDelegationTo(node address.Address) math.Ndau {
	out := math.Ndau(0)
	for delegateS := range s.Delegates[node.String()] {
		out += s.Accounts[delegateS].Balance
	}
	return out
}
