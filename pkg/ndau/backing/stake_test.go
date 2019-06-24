package backing

import (
	"testing"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

var (
	rules     address.Address
	primary   address.Address
	costaker1 address.Address
	costaker2 address.Address
)

func init() {
	rules = randAddress()
	primary = randAddress()
	costaker1 = randAddress()
	costaker2 = randAddress()
}

type tstake struct {
	qty     math.Ndau
	target  address.Address
	stakeTo address.Address
}

func (ts tstake) apply(t *testing.T, state *State) {
	t.Helper()

	targetAcct := state.Accounts[ts.target.String()]
	targetAcct.Balance += ts.qty

	isPrimary := ts.stakeTo == rules

	if isPrimary {
		ps := targetAcct.PrimaryStake(rules)
		require.Nil(t, ps, "target cannot have more than 1 primary stake to a rules account")
	}

	hold := Hold{
		Qty: ts.qty,
		Stake: &StakeData{
			RulesAcct: rules,
			StakeTo:   ts.stakeTo,
		},
	}
	targetAcct.Holds = append(targetAcct.Holds, hold)
	state.Accounts[ts.target.String()] = targetAcct

	if isPrimary {
		rulesAcct := state.Accounts[rules.String()]
		rulesAcct.StakeRules.Inbound[ts.target.String()]++
		state.Accounts[rules.String()] = rulesAcct
	} else {
		stakeToAcct := state.Accounts[ts.stakeTo.String()]
		rulesCostakers := stakeToAcct.Costakers[rules.String()]
		if rulesCostakers == nil {
			rulesCostakers = make(map[string]uint64)
		}
		rulesCostakers[ts.target.String()]++
		if stakeToAcct.Costakers == nil {
			stakeToAcct.Costakers = make(map[string]map[string]uint64)
		}
		stakeToAcct.Costakers[rules.String()] = rulesCostakers
		state.Accounts[ts.stakeTo.String()] = stakeToAcct
	}
}

func TestState_TotalStake(t *testing.T) {
	tests := []struct {
		name    string
		stakes  []tstake
		target  address.Address
		wantQty math.Ndau
	}{
		{
			"self-stake with no primary stake",
			[]tstake{{2, primary, primary}},
			primary,
			2,
		},
		{
			"single primary stake",
			[]tstake{{2, primary, rules}},
			primary,
			2,
		},
		{
			"primary stake with additional self-stake",
			[]tstake{
				{3, primary, primary},
				{2, primary, rules},
			},
			primary,
			5,
		},
		{
			"costaker",
			[]tstake{
				{3, costaker1, primary},
				{2, primary, rules},
			},
			costaker1,
			3,
		},
		{
			"costaker with another costaker",
			[]tstake{
				{3, costaker1, primary},
				{5, costaker2, primary},
				{2, primary, rules},
			},
			costaker1,
			3,
		},
		{
			"costaker with multiple stake and another costaker",
			[]tstake{
				{3, costaker1, primary},
				{7, costaker1, primary},
				{5, costaker2, primary},
				{2, primary, rules},
			},
			costaker1,
			10,
		},
		{
			"costaker with multiple stake and multiple primary stake and another costaker",
			[]tstake{
				{3, costaker1, primary},
				{7, costaker1, primary},
				{5, costaker2, primary},
				{11, primary, primary},
				{2, primary, rules},
			},
			costaker1,
			10,
		},
		{
			"costaker with primary self-stake",
			[]tstake{
				{3, costaker1, primary},
				{5, primary, primary},
				{2, primary, rules},
			},
			costaker1,
			3,
		},
		{
			"with costaker",
			[]tstake{
				{3, costaker1, primary},
				{2, primary, rules},
			},
			primary,
			2,
		},
		{
			"with two costakers",
			[]tstake{
				{3, costaker1, primary},
				{5, costaker2, primary},
				{2, primary, rules},
			},
			primary,
			2,
		},
		{
			"with costaker and self-stake",
			[]tstake{
				{3, costaker1, primary},
				{5, primary, primary},
				{2, primary, rules},
			},
			primary,
			7,
		},
		{
			"ignore external",
			[]tstake{
				{15, costaker2, costaker1},
				{3, costaker1, primary},
				{5, primary, primary},
				{2, primary, rules},
			},
			primary,
			7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up the state
			state := State{}
			state.Init(nil)

			state.Accounts[rules.String()] = AccountData{
				StakeRules: &StakeRules{
					Inbound: make(map[string]uint64),
				},
			}

			for _, stake := range tt.stakes {
				stake.apply(t, &state)
			}

			gotQty := state.TotalStake(tt.target, primary, rules)
			require.Equal(t, tt.wantQty, gotQty)
		})
	}
}

func TestState_AggregateStake(t *testing.T) {
	tests := []struct {
		name    string
		stakes  []tstake
		wantQty math.Ndau
		wantNil bool
	}{
		{
			"self-stake with no primary stake",
			[]tstake{{2, primary, primary}},
			0,
			true,
		},
		{
			"single primary stake",
			[]tstake{{2, primary, rules}},
			2,
			false,
		},
		{
			"primary stake with additional self-stake",
			[]tstake{
				{3, primary, primary},
				{2, primary, rules},
			},
			5,
			false,
		},
		{
			"costaker",
			[]tstake{
				{3, costaker1, primary},
				{2, primary, rules},
			},
			5,
			false,
		},
		{
			"costaker with another costaker",
			[]tstake{
				{3, costaker1, primary},
				{5, costaker2, primary},
				{2, primary, rules},
			},
			10,
			false,
		},
		{
			"costaker with multiple stake and another costaker",
			[]tstake{
				{3, costaker1, primary},
				{7, costaker1, primary},
				{5, costaker2, primary},
				{2, primary, rules},
			},
			17,
			false,
		},
		{
			"costaker with multiple stake and multiple primary stake and another costaker",
			[]tstake{
				{3, costaker1, primary},
				{7, costaker1, primary},
				{5, costaker2, primary},
				{11, primary, primary},
				{2, primary, rules},
			},
			28,
			false,
		},
		{
			"costaker with primary self-stake",
			[]tstake{
				{3, costaker1, primary},
				{5, primary, primary},
				{2, primary, rules},
			},
			10,
			false,
		},
		{
			"with costaker",
			[]tstake{
				{3, costaker1, primary},
				{2, primary, rules},
			},
			5,
			false,
		},
		{
			"with two costakers",
			[]tstake{
				{3, costaker1, primary},
				{5, costaker2, primary},
				{2, primary, rules},
			},
			10,
			false,
		},
		{
			"with costaker and self-stake",
			[]tstake{
				{3, costaker1, primary},
				{5, primary, primary},
				{2, primary, rules},
			},
			10,
			false,
		},
		{
			"ignore external",
			[]tstake{
				{15, costaker2, costaker1},
				{3, costaker1, primary},
				{5, primary, primary},
				{2, primary, rules},
			},
			10,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up the state
			state := State{}
			state.Init(nil)

			state.Accounts[rules.String()] = AccountData{
				StakeRules: &StakeRules{
					Inbound: make(map[string]uint64),
				},
			}

			for _, stake := range tt.stakes {
				stake.apply(t, &state)
			}

			got := state.AggregateStake(primary, rules)
			if tt.wantNil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.Equal(t, tt.wantQty, got.Qty)
			}
		})
	}
}
