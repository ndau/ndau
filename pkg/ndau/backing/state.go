package backing

import (
	nt "github.com/attic-labs/noms/go/types"
	meta "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// State is primarily a set of accounts
//nomsify State
type State struct {
	Accounts map[string]AccountData
	// Delegates is a map of strings to a set of strings
	// All strings are addresses
	// the keys are the delegated nodes
	// the values are the addresses of the accounts which those nodes must
	// compute
	Delegates map[string]map[string]struct{}
	// Nodes keeps track of the validator and verifier node stakes.
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
	// TotalSIB is the sum of all burned SIB
	TotalSIB math.Ndau
	// TotalFees is the sum of all burned tx fees
	TotalFees math.Ndau
	// These prices are preserved here just to assist downstream consumers
	// have more trust in the SIB calculations.
	MarketPrice pricecurve.Nanocent
	TargetPrice pricecurve.Nanocent
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
	defaultSettlementPeriod math.Duration,
) (AccountData, bool) {
	data, hasAccount := s.Accounts[address.String()]
	if !hasAccount {
		data = NewAccountData(blockTime, defaultSettlementPeriod)
	}
	return data, hasAccount
}

// Stake updates the state to handle staking an account to another
func (s *State) Stake(targetA, nodeA address.Address) error {
	nodeS := nodeA.String()
	node, isNode := s.Nodes[nodeS]
	// logically, the operation I want in this if is nxor, but go doesn't
	// define that for booleans, because reasons
	if (targetA == nodeA) == isNode {
		if isNode {
			return errors.New("cannot re-self-stake")
		}
		return errors.New("node is not already a node; can't stake to it")
	}

	target := s.Accounts[targetA.String()]
	if isNode {
		// targetA != nodeA
		node.Costake(targetA, target.Balance)
	} else {
		// targetA == nodeA
		node = NewNode(targetA, target.Balance)
	}

	s.Nodes[nodeS] = node
	return nil
}

// Unstake updates the state to handle unstaking an account
func (s *State) Unstake(targetA address.Address) {
	target, exists := s.Accounts[targetA.String()]
	if !exists {
		return
	}
	if target.Stake == nil {
		return
	}
	nodeA := target.Stake.Address
	target.Stake = nil
	s.Accounts[targetA.String()] = target

	node, isNode := s.Nodes[nodeA.String()]
	if isNode {
		// targetA != nodeA
		node.Unstake(targetA)
		s.Nodes[nodeA.String()] = node
	}
}

// GetCostakers returns the list of costakers associated with a node
func (s *State) GetCostakers(nodeA address.Address) []AccountData {
	node, isNode := s.Nodes[nodeA.String()]
	if !isNode {
		return nil
	}

	out := make([]AccountData, 0, len(node.Costakers))
	for costaker := range node.Costakers {
		ad, hasAccount := s.Accounts[costaker]
		if hasAccount {
			out = append(out, ad)
		}
	}
	return out
}

// PayReward updates the state of the target address to add the given qty of ndau, following
// the link to any specified rewards target. If the rewards target account does not previously exist,
// it will be created. Returns a list of accounts whose state was updated in some way.
// Note that if the reward is redirected, it is still the original account whose lastEAIUpdate time
// is changed.
// If isEAI we change WAA only if it's not the same account (per the rules of EAI).
// If it's redirected we change WAA for the target account and do nothing to the source.
// If it's not redirected we change WAA only if it's not EAI.
// After updating the balance in an account, this also updates currency seat
// information for that account.
func (s *State) PayReward(
	srcAddress address.Address,
	reward math.Ndau,
	blockTime math.Timestamp,
	defaultSettlementPeriod math.Duration,
	isEAI bool,
) ([]address.Address, error) {
	var err error
	srcAccount, _ := s.GetAccount(srcAddress, blockTime, defaultSettlementPeriod)

	if srcAccount.RewardsTarget != nil {
		// rewards are being redirected, so get the target account
		tgtAddress := *(srcAccount.RewardsTarget)
		tgtAccount, _ := s.GetAccount(tgtAddress, blockTime, defaultSettlementPeriod)
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
