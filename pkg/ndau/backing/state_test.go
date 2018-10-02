package backing

import (
	"fmt"
	"sort"
	"testing"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

func randomState(t *testing.T, n int, qty types.Ndau, rewardsTarget bool) State {
	s := State{}
	s.Init(nil)
	for i := 0; i < n; i++ {
		data, _ := generateAccount(t, qty, false, false)
		if rewardsTarget {
			addr := randAddress()
			data.RewardsTarget = &addr
		} else {
			data.RewardsTarget = nil
		}
		addr := randAddress()
		s.Accounts[addr.String()] = data
	}
	fmt.Println(s)
	return s
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
		isEAI    = true
		notEAI   = false
		isRedir  = true
		noRedir  = false
		chkWAA   = true
		nochkWAA = false
		chkEAI   = true
		nochkEAI = false
	)

	type args struct {
		ix     int
		reward types.Ndau
	}
	tests := []struct {
		name    string
		numaddr int
		qty     types.Ndau
		eai     bool
		redir   bool
		args    args
		want    int
		amount  types.Ndau
		ckWAA   bool
		ckEAI   bool
		wantErr bool
	}{
		{"simple", 1, 0, notEAI, noRedir, args{0, 1234}, 0, 1234, chkWAA, nochkEAI, false},
		{"nonzero simple", 1, 1234, notEAI, noRedir, args{0, 1}, 0, 1235, chkWAA, nochkEAI, false},
		{"redirect", 1, 1234, notEAI, isRedir, args{0, 1}, 0, 1, chkWAA, nochkEAI, false},
		{"eai", 1, 1234, isEAI, isRedir, args{0, 1}, 0, 1, chkWAA, chkEAI, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := randomState(t, tt.numaddr, tt.qty, tt.redir)
			addr := getAddr(getAccountID(s, tt.args.ix))
			wantid := getAccountID(s, tt.want)
			if tt.redir {
				wantid = (*s.Accounts[wantid].RewardsTarget).String()
			}
			fmt.Println(addr, wantid)
			for a := range s.Accounts {
				fmt.Println("Acct: ", a)
			}
			blockTime := randTimestamp()
			got, err := s.PayReward(addr, tt.args.reward, blockTime, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("State.PayReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != wantid {
				t.Errorf("State.PayReward() = %v, want %v", got, wantid)
			}
			acct := s.Accounts[got.String()]
			if acct.Balance != tt.amount {
				t.Errorf("State[%s].amount = %v, want %v", got, acct.Balance, tt.amount)
			}
			if tt.ckEAI && acct.LastEAIUpdate != blockTime {
				t.Errorf("State[%s].LastEAIUpdate = %v, want %v", got, acct.LastEAIUpdate, blockTime)
			}
			if tt.ckWAA && acct.LastWAAUpdate != blockTime {
				t.Errorf("State[%s].LastWAAUpdate = %v, want %v", got, acct.LastWAAUpdate, blockTime)
			}
		})
	}
}
