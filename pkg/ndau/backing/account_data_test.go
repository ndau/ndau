package backing

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/oneiro-ndev/signature/pkg/signature"

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
					// transfer key may not be equal if algorithm pointers are unequal
					require.Equal(t, account.TransferKey.Bytes(), recoveredAccount.TransferKey.Bytes())
					require.Equal(t, account.RewardsTarget, recoveredAccount.RewardsTarget)
					require.Equal(t, account.IncomingRewardsFrom, recoveredAccount.IncomingRewardsFrom)
					require.Equal(t, account.DelegationNode, recoveredAccount.DelegationNode)
					require.Equal(t, account.Lock, recoveredAccount.Lock)
					require.Equal(t, account.Stake, recoveredAccount.Stake)
					require.Equal(t, account.LastWAAUpdate, recoveredAccount.LastWAAUpdate)
					require.Equal(t, account.WeightedAverageAge, recoveredAccount.WeightedAverageAge)
					require.Equal(t, account.Sequence, recoveredAccount.Sequence)
					require.Equal(t, account.Settlements, recoveredAccount.Settlements)
					require.Equal(t, account.SettlementSettings, recoveredAccount.SettlementSettings)
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

func randKey() *signature.PublicKey {
	public, _, err := signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
	}
	return &public
}

func randAddress() address.Address {
	key := randKey()
	addr, _ := address.Generate(
		address.KindUser,
		key.Bytes(),
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
		SettlementSettings: generateEscrowSettings(randBool()),
	}
	if randBool() {
		addr := randAddress()
		ad.RewardsTarget = &addr
	}
	if randBool() {
		addr := randAddress()
		ad.DelegationNode = &addr
	}
	for i := 0; i < 5; i++ {
		ad.IncomingRewardsFrom = append(ad.IncomingRewardsFrom, randAddress())
	}
	if hasLock {
		ad.Lock = generateLock(randBool())
	}
	if hasStake {
		ad.Stake = generateStake()
	}
	qtyEscrows := rand.Intn(10)
	for i := 0; i < qtyEscrows; i++ {
		ad.Settlements = append(ad.Settlements, generateEscrow())
	}
	return ad, name
}

func generateLock(notified bool) *Lock {
	l := &Lock{
		NoticePeriod: randDuration(),
	}
	if randBool() {
		ts := randTimestamp()
		l.UnlocksOn = &ts
	}
	return l
}

func generateStake() *Stake {
	return &Stake{
		Point:   randTimestamp(),
		Address: randAddress(),
	}
}

func generateEscrow() Settlement {
	return Settlement{
		Qty:    randNdau(),
		Expiry: randTimestamp(),
	}
}

func generateEscrowSettings(changing bool) SettlementSettings {
	es := SettlementSettings{
		Period: randDuration(),
	}
	if changing {
		ts := randTimestamp()
		es.ChangesAt = &ts
		n := randDuration()
		es.Next = &n
	}
	return es
}

func TestUpdateEscrow(t *testing.T) {
	// create fixture
	const baseNdau = 100
	baseTimestamp, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	acct, _ := generateAccount(t, baseNdau, false, false)
	// get rid of the random escrows generated by generateAccount
	acct.Settlements = nil
	const qtyEscrows = 20
	for i := 1; i <= qtyEscrows/2; i++ {
		acct.Settlements = append(
			acct.Settlements,
			Settlement{
				Qty:    math.Ndau(i),
				Expiry: math.Timestamp(baseTimestamp.Sub(math.Duration(i))),
			},
		)
		acct.Settlements = append(
			acct.Settlements,
			Settlement{
				Qty:    math.Ndau(i),
				Expiry: math.Timestamp(baseTimestamp.Add(math.Duration(i))),
			},
		)
	}

	acct.UpdateSettlement(baseTimestamp)

	// half of the escrows are after the base timestamp
	require.Equal(t, qtyEscrows/2, len(acct.Settlements))

	expectedNdau := baseNdau
	for i := 1; i <= qtyEscrows/2; i++ {
		expectedNdau += i
	}
	require.Equal(t, math.Ndau(expectedNdau), acct.Balance)
}

func TestUpdateEscrowUpdatesPendingPeriodChange(t *testing.T) {
	// create fixture
	const baseNdau = 100
	acct, _ := generateAccount(t, baseNdau, false, false)
	chTs := randTimestamp()
	chD := randDuration()
	acct.SettlementSettings.Period = randDuration()
	acct.SettlementSettings.ChangesAt = &chTs
	acct.SettlementSettings.Next = &chD

	acct.UpdateSettlement(chTs)

	require.Equal(t, chD, acct.SettlementSettings.Period)
	require.Nil(t, acct.SettlementSettings.Next)
	require.Nil(t, acct.SettlementSettings.ChangesAt)
}

func TestUpdateEscrowPersistsPendingPeriodChange(t *testing.T) {
	// create fixture
	const baseNdau = 100
	acct, _ := generateAccount(t, baseNdau, false, false)
	stD := randDuration()
	chTs := randTimestamp()
	chD := randDuration()
	acct.SettlementSettings.Period = stD
	acct.SettlementSettings.ChangesAt = &chTs
	acct.SettlementSettings.Next = &chD

	acct.UpdateSettlement(chTs.Sub(math.Duration(1)))

	require.Equal(t, stD, acct.SettlementSettings.Period)
	require.Equal(t, &chD, acct.SettlementSettings.Next)
	require.Equal(t, &chTs, acct.SettlementSettings.ChangesAt)
}
