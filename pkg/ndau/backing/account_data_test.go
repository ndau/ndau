package backing

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/attic-labs/noms/go/chunks"
	"github.com/attic-labs/noms/go/marshal"
	"github.com/attic-labs/noms/go/spec"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
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
					require.Equal(t, len(account.ValidationKeys), len(recoveredAccount.ValidationKeys))
					for idx := range account.ValidationKeys {
						// transfer key may not be equal if algorithm pointers are unequal
						require.Equal(t, account.ValidationKeys[idx].KeyBytes(), recoveredAccount.ValidationKeys[idx].KeyBytes())
					}
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
					// validaation scripts of nil or 0 length are equivalent
					require.Equal(t, len(account.ValidationScript), len(recoveredAccount.ValidationScript))
					if len(account.ValidationScript) > 0 {
						require.Equal(t, account.ValidationScript, recoveredAccount.ValidationScript)
					}
					require.Equal(t, account.UncreditedEAI, recoveredAccount.UncreditedEAI)
					require.Equal(t, account.SidechainPayments, recoveredAccount.SidechainPayments)
					require.Equal(t, account.CurrencySeatDate, recoveredAccount.CurrencySeatDate)
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
		key.KeyBytes(),
	)
	return addr
}

func generateAccount(t *testing.T, balance math.Ndau, hasLock, hasStake bool) (AccountData, string) {
	t.Helper()
	name := fmt.Sprintf("<Account (bal: %d; lock: %v; stake: %v)>", int64(balance), hasLock, hasStake)
	ad := AccountData{
		Balance:            balance,
		LastWAAUpdate:      randTimestamp(),
		WeightedAverageAge: randDuration(),
		Sequence:           rand.Uint64(),
		SettlementSettings: generateEscrowSettings(randBool()),
		UncreditedEAI:      randNdau(),
		SidechainPayments:  make(map[string]struct{}),
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
		key := randKey()
		ad.ValidationKeys = append(ad.ValidationKeys, *key)
		ad.IncomingRewardsFrom = append(ad.IncomingRewardsFrom, randAddress())
	}
	if hasLock {
		ad.Lock = generateLock(randBool())
		// verify that account roundtrips include non-0 lock bonuses
		// t.Log("generated lock bonus:", ad.Lock.Bonus)
	}
	if hasStake {
		ad.Stake = generateStake()
	}
	qtyEscrows := rand.Intn(10)
	for i := 0; i < qtyEscrows; i++ {
		ad.Settlements = append(ad.Settlements, generateEscrow())
	}
	if randBool() {
		ad.ValidationScript = make([]byte, 20)
		rand.Read(ad.ValidationScript)
	}
	qtySidechainPayments := rand.Intn(4)
	for i := 0; i < qtySidechainPayments; i++ {
		ad.SidechainPayments[randAddress().String()] = struct{}{}
	}
	if randBool() {
		csd := randTimestamp()
		ad.CurrencySeatDate = &csd
	}
	return ad, name
}

func generateLock(notified bool) *Lock {
	l := NewLock(randDuration(), eai.DefaultLockBonusEAI)
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
		acct.Balance += math.Ndau(2 * i)
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

	acct.UpdateSettlements(baseTimestamp)
	available, err := acct.AvailableBalance()
	require.NoError(t, err)

	// half of the escrows are after the base timestamp
	require.Equal(t, qtyEscrows/2, len(acct.Settlements))

	expectedNdau := baseNdau
	for i := 1; i <= qtyEscrows/2; i++ {
		expectedNdau += i
	}

	require.Equal(t, math.Ndau(expectedNdau), available)
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

	acct.UpdateSettlements(chTs)

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

	acct.UpdateSettlements(chTs.Sub(math.Duration(1)))

	require.Equal(t, stD, acct.SettlementSettings.Period)
	require.Equal(t, &chD, acct.SettlementSettings.Next)
	require.Equal(t, &chTs, acct.SettlementSettings.ChangesAt)
}

func TestAccountData_ValidateSignatures(t *testing.T) {
	data := make([]byte, 512)
	_, err := rand.Read(data)
	require.NoError(t, err)

	const keypairQty = 12
	type keypairsig struct {
		public    signature.PublicKey
		private   signature.PrivateKey
		signature signature.Signature
	}
	keypairs := make([]keypairsig, 0, keypairQty)
	for i := 0; i < keypairQty; i++ {
		public, private, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)

		kp := keypairsig{public: public, private: private}
		kp.signature = private.Sign(data)
		keypairs = append(keypairs, kp)
	}

	kpPublic := func(idxs ...int) []signature.PublicKey {
		keys := make([]signature.PublicKey, 0, len(idxs))
		for _, idx := range idxs {
			keys = append(keys, keypairs[idx].public)
		}
		return keys
	}

	kpSignature := func(idxs ...int) []signature.Signature {
		keys := make([]signature.Signature, 0, len(idxs))
		for _, idx := range idxs {
			keys = append(keys, keypairs[idx].signature)
		}
		return keys
	}

	tests := []struct {
		name  string
		keys  []signature.PublicKey
		sigs  []signature.Signature
		want  bool
		want1 *bitset256.Bitset256
	}{
		{
			"1 valid",
			kpPublic(0),
			kpSignature(0),
			true, bitset256.New(0),
		},
		{
			"1 invalid",
			kpPublic(1),
			kpSignature(2),
			false, bitset256.New(0),
		},
		{
			"2 valid out of order",
			kpPublic(3, 4),
			kpSignature(4, 3),
			true, bitset256.New(0, 1),
		},
		{
			"any invalid sig invalidates all",
			kpPublic(5, 6, 7),
			kpSignature(7, 3, 6),
			false, bitset256.New(1),
		},
		{
			"valid subset is valid",
			kpPublic(8, 9, 10, 11),
			kpSignature(10, 8),
			true, bitset256.New(0, 2),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := &AccountData{
				ValidationKeys: tt.keys,
			}
			got, got1 := ad.ValidateSignatures(data, tt.sigs)
			if got != tt.want {
				t.Errorf("AccountData.ValidateSignatures() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("AccountData.ValidateSignatures() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func newTestValueStore() *nt.ValueStore {
	ts := &chunks.TestStorage{}
	return nt.NewValueStore(ts.NewView())
}

var vval nt.Value

func BenchmarkMarshalNomsAccountData(b *testing.B) {
	v := AccountData{}

	vrw := newTestValueStore()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vval, _ = v.MarshalNoms(vrw)
	}
}

var ad AccountData

func BenchmarkUnmarshalNomsAccountData(b *testing.B) {
	v := AccountData{}

	vrw := newTestValueStore()
	vval, _ = v.MarshalNoms(vrw)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ad.UnmarshalNoms(vval)
	}
}
