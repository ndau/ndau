package backing

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/attic-labs/noms/go/marshal"
	"github.com/attic-labs/noms/go/spec"
	"github.com/stretchr/testify/require"

	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	address "github.com/oneiro-ndev/ndaunode/pkg/node.address"
)

func alphaOf(in string) (out string) {
	for _, ch := range in {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			out += string(ch)
		}
	}
	return
}

func TestAccountDataRoundTrip(t *testing.T) {
	// set up noms
	sp, err := spec.ForDatabase("mem")
	require.NoError(t, err)
	db := sp.GetDatabase()

	for _, hasLock := range []bool{false, true} {
		for _, hasStake := range []bool{false, true} {
			account, name := generateAccount(t, randNdau(), hasLock, hasStake)
			ds := db.GetDataset(alphaOf(name))

			t.Run(name, func(t *testing.T) {
				// write to noms
				nomsAccount, err := marshal.Marshal(db, account)
				require.NoError(t, err)
				ds, err := db.CommitValue(ds, nomsAccount)
				require.NoError(t, err)
				// recover from noms
				recoveredNomsAccount, hasHead := ds.MaybeHeadValue()
				require.True(t, hasHead)
				var recoveredAccount AccountData
				err = marshal.Unmarshal(recoveredNomsAccount, &recoveredAccount)
				require.NoError(t, err)

				require.True(t, reflect.DeepEqual(account, recoveredAccount))
			})
		}
	}
}

func randBool() bool {
	return rand.Intn(1) == 1
}

func randNdau() math.Ndau {
	return math.Ndau(rand.Int63n(2500 * constants.QuantaPerUnit))
}

func randDuration() math.Duration {
	return math.Duration(rand.Int63n(3 * math.Year))
}

func randTimestamp() math.Timestamp {
	return math.Timestamp(rand.Int63n(5 * math.Year))
}

func generateAccount(t *testing.T, balance math.Ndau, hasLock, hasStake bool) (AccountData, string) {
	t.Helper()
	name := fmt.Sprintf("<Account (bal: %d; lock: %v; stake: %v)>", int64(balance), hasLock, hasStake)
	ad := AccountData{
		Balance:        balance,
		UpdatePoint:    randTimestamp(),
		Sequence:       rand.Uint64(),
		EscrowSettings: generateEscrowSettings(randBool()),
	}
	if hasLock {
		ad.Lock = generateLock(randBool())
	}
	if hasStake {
		ad.Stake = generateStake()
	}
	qtyEscrows := rand.Intn(10)
	for i := 0; i < qtyEscrows; i++ {
		ad.Escrows = append(ad.Escrows, generateEscrow())
	}
	return ad, name
}

func generateLock(notified bool) *Lock {
	l := &Lock{
		Duration: randDuration(),
	}
	if randBool() {
		*l.NotifiedOn = randTimestamp()
	}
	return l
}

func generateStake() *Stake {
	return &Stake{
		Point:   randTimestamp(),
		Address: address.Address(make([]byte, 32)),
	}
}

func generateEscrow() Escrow {
	return Escrow{
		Qty:    randNdau(),
		Expiry: randTimestamp(),
	}
}

func generateEscrowSettings(changing bool) EscrowSettings {
	es := EscrowSettings{
		Duration: randDuration(),
	}
	if changing {
		*es.ChangesAt = randTimestamp()
		*es.Next = randDuration()
	}
	return es
}
