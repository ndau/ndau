package backing

import (
	"sort"
	"testing"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

func randomState(t *testing.T, qty types.Ndau, rewardsTarget bool) (address.Address, State) {
	s := State{}
	s.Init(nil)
	data, _ := generateAccount(t, qty, false, false)
	if rewardsTarget {
		tgtaddr := randAddress()
		// we want an empty target account that already exists so that its date fields are set
		tgtdata, _ := generateAccount(t, 0, false, false)
		s.Accounts[tgtaddr.String()] = tgtdata
		data.RewardsTarget = &tgtaddr
	} else {
		data.RewardsTarget = nil
	}
	addr := randAddress()
	s.Accounts[addr.String()] = data
	return addr, s
}

// getAccount sorts the list of account indices and returns the nth item;
// it panics if there aren't n entries in the list
func getAccountID(s State, n int) string {
	sa := make([]string, 0)
	for k := range s.Accounts {
		sa = append(sa, k)
	}
	sort.Sort(sort.StringSlice(sa))
	return sa[n]
}

func getAddr(id string) address.Address {
	a, _ := address.Validate(id)
	return a
}

func TestState_PayReward(t *testing.T) {
	const (
		isEAI   = true
		notEAI  = false
		isRedir = true
		noRedir = false
		chgW    = true
		nochgW  = false
		chgE    = true
		nochgE  = false
	)

	// test account data for each returned value
	type ta struct {
		balance types.Ndau
		chgWAA  bool
		chgEAI  bool
	}

	tests := []struct {
		name    string
		qty     types.Ndau
		eai     bool
		redir   bool
		reward  types.Ndau
		resp    []ta
		wantErr bool
	}{
		// reward nothing, do almost nothing (LastWAAUpdate should be fresh)
		{"simple", 0, notEAI, noRedir, 1234, []ta{ta{1234, chgW, nochgE}}, false},
		// reward something, update Balance and LastWAAUpdate
		{"nonzero simple", 1234, notEAI, noRedir, 1, []ta{ta{1235, chgW, nochgE}}, false},
		// reward EAI, expect WAA not to change
		{"nonzero simple EAI", 1234, isEAI, noRedir, 1, []ta{ta{1235, nochgW, chgE}}, false},
		// redirect non-EAI, source unchanged, target changes WAA
		{"redirect nonEAI", 1234, notEAI, isRedir, 1, []ta{ta{1, chgW, nochgE}}, false},
		// redirect EAI, source changes EAI, target changes WAA
		{"redirect EAI", 1234, isEAI, isRedir, 1, []ta{ta{1234, nochgW, chgE}, ta{1, chgW, nochgE}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, s := randomState(t, tt.qty, tt.redir)
			var wantids []string
			switch {
			case tt.redir == false:
				wantids = []string{addr.String()}
			case tt.redir == true && tt.eai == false:
				wantids = []string{(*s.Accounts[addr.String()].RewardsTarget).String()}
			case tt.redir == true && tt.eai == true:
				wantids = []string{addr.String(), (*s.Accounts[addr.String()].RewardsTarget).String()}
			}
			blockTime := randTimestamp()
			got, err := s.PayReward(addr, tt.reward, blockTime, 0, tt.eai, tt.eai)
			if (err != nil) != tt.wantErr {
				t.Errorf("State.PayReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(wantids) {
				t.Errorf("State.PayReward returned wrong length result: %v but we wanted %v", got, wantids)
				return
			}
			for i, a := range got {
				if a.String() != wantids[i] {
					t.Errorf("State.PayReward() returned wrong ID = %v, want %v", a, wantids[i])
				}
				acctid := a.String()
				acct := s.Accounts[acctid]
				if acct.Balance != tt.resp[i].balance {
					t.Errorf("%d) State[%s].Balance = %v, want %v", i, acctid, acct.Balance, tt.resp[i].balance)
				}
				if tt.resp[i].chgEAI == false && acct.LastEAIUpdate == blockTime {
					t.Errorf("%d) State[%s].LastEAIUpdate was updated but shouldn't have been", i, acctid)
				}
				if tt.resp[i].chgWAA == false && acct.LastWAAUpdate == blockTime {
					t.Errorf("%d) State[%s].LastWAAUpdate was updated but shouldn't have been", i, acctid)
				}
			}
		})
	}
}
