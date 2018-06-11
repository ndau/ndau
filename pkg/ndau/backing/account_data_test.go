package backing

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/attic-labs/noms/go/marshal"
	"github.com/attic-labs/noms/go/spec"
	"github.com/stretchr/testify/require"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

func alphaOf(in string) (out string) {
	for _, ch := range in {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
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
			// generated accounts have a bunch of random fields
			// to give confidence that a single test run tests them all,
			// we just run this configuration of subtests a bunch of times
			for i := 0; i < 10; i++ {
				account, name := generateAccount(t, randNdau(), hasLock, hasStake)
				name = fmt.Sprintf("%s (%d)", name, i)
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

					// require equality for known fields by name so we know what's
					// unequal, if anything is
					require.Equal(t, account.Balance, recoveredAccount.Balance)
					require.Equal(t, account.TransferKey, recoveredAccount.TransferKey)
					require.Equal(t, account.RewardsTarget, recoveredAccount.RewardsTarget)
					require.Equal(t, account.DelegationNode, recoveredAccount.DelegationNode)
					require.Equal(t, account.Lock, recoveredAccount.Lock)
					require.Equal(t, account.Stake, recoveredAccount.Stake)
					require.Equal(t, account.LastWAAUpdate, recoveredAccount.LastWAAUpdate)
					require.Equal(t, account.WeightedAverageAge, recoveredAccount.WeightedAverageAge)
					require.Equal(t, account.Sequence, recoveredAccount.Sequence)
					require.Equal(t, account.Escrows, recoveredAccount.Escrows)
					require.Equal(t, account.EscrowSettings, recoveredAccount.EscrowSettings)
					// require deep equality just in case we add something later
					require.True(t, reflect.DeepEqual(account, recoveredAccount))
				})
			}
		}
	}
}

func randBool() bool {
	return rand.Intn(2) == 0
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

func randKey() []byte {
	key := make([]byte, address.MinDataLength)
	rand.Read(key)
	return key
}

func randAddress() string {
	addr, _ := address.Generate(
		address.KindUser,
		randKey(),
	)
	return addr
}

func generateAccount(t *testing.T, balance math.Ndau, hasLock, hasStake bool) (AccountData, string) {
	t.Helper()
	name := fmt.Sprintf("<Account (bal: %d; lock: %v; stake: %v)>", int64(balance), hasLock, hasStake)
	ad := AccountData{
		Balance:            balance,
		TransferKey:        randKey(),
		LastWAAUpdate:      randTimestamp(),
		WeightedAverageAge: randDuration(),
		Sequence:           rand.Uint64(),
		EscrowSettings:     generateEscrowSettings(randBool()),
	}
	if randBool() {
		ad.RewardsTarget = randAddress()
	}
	if randBool() {
		ad.DelegationNode = randAddress()
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
		NoticePeriod: randDuration(),
	}
	if randBool() {
		ts := randTimestamp()
		l.NotifiedOn = &ts
	}
	return l
}

func generateStake() *Stake {
	return &Stake{
		Point:   randTimestamp(),
		Address: randAddress(),
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
		ts := randTimestamp()
		es.ChangesAt = &ts
		n := randDuration()
		es.Next = &n
	}
	return es
}
