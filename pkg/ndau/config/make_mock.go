package config

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/oneiro-ndev/chaos/pkg/tool"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/ndaumath/pkg/signed"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tinylib/msgp/msgp"
	"golang.org/x/crypto/ed25519"
)

// we redefine the system namespace here instead of depending on the actual
// value from the chaos chain. It doesn't particularly matter whether we get
// it right for the purposes of this test, and it simplifies the dependency
// management story considerably
var system = []byte("system")

// ensure that the parents of a given path exist
func ensureDir(path string) error {
	parent := filepath.Dir(path)
	return os.MkdirAll(parent, 0700)
}

// MockAssociated tracks associated data which goes with the mocks.
//
// In particular, it's used for tests. For example, we mock up some
// public/private keypairs for the ReleaseFromEndowment transaction.
// The public halves of those keys are written into the mock file,
// but the private halves are communicated to the test suite by means
// of the MockAssociated struct.
type MockAssociated map[string]interface{}

// MakeChaosMock loads mock system variables into the Chaos chain
func MakeChaosMock(config *Config) (MockAssociated, error) {
	if config.ChaosAddress == "" {
		return nil, errors.New("chaos address not set in config")
	}

	bpcPublic, bpcPrivate, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	svi := makeMockSVI(bpcPublic, false)
	chaosMock, ma, sviKey := makeMockChaos(bpcPublic, svi, false)
	config.SystemVariableIndirect = *sviKey

	// we don't iterate through the chaosMock map's outer layer,
	// because we don't have appropriate private keys for everything.
	// instead, we pick out the BPC namespace, set those, and then
	// hope for the best.
	node := client.NewHTTP(config.ChaosAddress, "/websocket")
	bpcMap := chaosMock[string([]byte(bpcPublic))]
	fmt.Fprintf(os.Stderr, "Writing mocks to chaos at %s\n", config.ChaosAddress)
	fmt.Fprintf(os.Stderr, " (bpc ns: %s)\n", base64.RawStdEncoding.EncodeToString([]byte(bpcPublic)))
	for keyString, valB := range bpcMap {
		fmt.Fprintf(os.Stderr, "  %q... ", keyString)
		key := []byte(keyString)
		err = tool.SetRawCommit(
			node, key, valB, []byte(bpcPublic), []byte(bpcPrivate),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: %s\n", err.Error())
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to set key %s", keyString))
		}
		fmt.Fprintf(os.Stderr, "DONE\n")
	}

	// blank the UseMock field--after setting the actual chain mocks,
	// we don't need the mock file anymore.
	config.UseMock = nil

	return ma, nil
}

// MakeMock creates a mock file with the specififed data.
//
// If `configPath == ""`, the config file is skipped. Otherwise,
// the config file at that path is created and directed to the
// mock file.
func MakeMock(configPath, mockPath string) (config *Config, ma MockAssociated, err error) {
	bpc, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, err
	}

	svi := makeMockSVI(bpc, true)
	mock, ma, sviKey := makeMockChaos(bpc, svi, true)

	// make the mock file
	err = mock.Dump(mockPath)
	if err != nil {
		return nil, nil, err
	}

	if configPath != "" {
		config, err = LoadDefault(configPath)
		if err != nil {
			return nil, nil, err
		}
		config.UseMock = &mockPath
		config.SystemVariableIndirect = *sviKey
		err = config.Dump(configPath)
	}

	return config, ma, err
}

// MakeTmpMock makes a mock config with temporary files.
//
// `tmpdir` is the location in which to store these files.
// If it is blank, they're stored in a system-defined location.
//
// As we don't keep track of these files, they'll persist until
// the system cleans them up. On most OSX and Linux systems, that
// happens after three days of disuse. We can get away with this
// because they're small.
func MakeTmpMock(tmpdir string) (config *Config, ma MockAssociated, err error) {
	configFile, err := ioutil.TempFile("", "config")
	if err != nil {
		return nil, nil, err
	}
	mockFile, err := ioutil.TempFile("", "mock")
	if err != nil {
		return nil, nil, err
	}
	return MakeMock(configFile.Name(), mockFile.Name())
}

// mock up some chaos data
func makeMockChaos(bpc []byte, svi msgp.Marshaler, testVars bool) (ChaosMock, MockAssociated, *NamespacedKey) {
	mock := make(ChaosMock)
	var sviKey NamespacedKey
	if testVars {

		mock.Sets(system, "one", wkt.String("system value one"))
		mock.Sets(system, "two", wkt.String("system value two"))

		mock.Sets(bpc, "one", wkt.String("bpc val one"))
		mock.Sets(bpc, "bar", wkt.String("baz"))
	}

	if svi != nil {
		mock.Sets(bpc, "svi", svi)
		sviKey = NewNamespacedKey(bpc, "svi")
	}

	var err error

	// prepare to return associated data
	ma := make(MockAssociated)

	// set RFE address
	// generate ownership keys
	ma[sv.ReleaseFromEndowmentOwnershipName], ma[sv.ReleaseFromEndowmentOwnershipPrivateName], err = signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
	}
	// now generate and set the address
	rfeOwnership := ma[sv.ReleaseFromEndowmentOwnershipName].(signature.PublicKey)
	rfeAddr, err := address.Generate(address.KindNdau, rfeOwnership.KeyBytes())
	if err != nil {
		panic(err)
	}
	mock.Sets(
		bpc,
		sv.ReleaseFromEndowmentAddressName,
		rfeAddr,
	)

	// set default rate tables
	mock.Sets(bpc, sv.UnlockedRateTableName, eai.DefaultUnlockedEAI)
	mock.Sets(bpc, sv.LockedRateTableName, eai.DefaultLockBonusEAI)

	// make default escrow duration
	ded := sv.DefaultSettlementDuration{Duration: math.Day * 15}
	mock.Sets(bpc, sv.DefaultSettlementDurationName, ded)

	// make default tx fee script
	// this one is very simple: unconditionally returns numeric 0
	// (base64 oAAgiA== if you'd like to decompile)
	mock.Sets(bpc, sv.TxFeeScriptName, wkt.Bytes([]byte{
		0xa0,
		0x00,
		0x20,
		0x88,
	}))

	// min stake for an account to be active
	mock.Sets(bpc, sv.MinStakeName, math.Ndau(1000*constants.QuantaPerUnit))

	// make default node goodness script
	// empty: returns the value on top of the stack
	// as goodness functions have the total stake on top of the stack,
	// that's actually not a terrible default
	// (base64 oACI if you'd like to decompile)
	mock.Sets(bpc, sv.NodeGoodnessFuncName, wkt.Bytes([]byte{
		0xa0,
		0x00,
		0x88,
	}))

	// make eai fee table
	mock.Sets(bpc, sv.EAIFeeTableName, makeMockEAIFeeTable())

	// set default min duration between node rewards nominations
	mock.Sets(
		bpc,
		sv.MinDurationBetweenNodeRewardNominationsName,
		math.Duration(1*math.Day),
	)

	// set nominate reward
	// generate ownership keys
	ma[sv.NominateNodeRewardOwnershipName], ma[sv.NominateNodeRewardOwnershipPrivateName], err = signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
	}
	// now generate and set the address
	nnrOwnership := ma[sv.NominateNodeRewardOwnershipName].(signature.PublicKey)
	nnrAddr, err := address.Generate(address.KindNdau, nnrOwnership.KeyBytes())
	if err != nil {
		panic(err)
	}
	mock.Sets(
		bpc,
		sv.NominateNodeRewardAddressName,
		nnrAddr,
	)

	// set node reward nomination timeout
	mock.Sets(
		bpc,
		sv.NodeRewardNominationTimeoutName,
		math.Duration(30*math.Second),
	)

	return mock, ma, &sviKey
}

// mock up an SVI Map using most of its features
func makeMockSVI(bpc []byte, testVars bool) SVIMap {
	svi := make(SVIMap)
	if testVars {
		svi.set("one", NewNamespacedKey(bpc, "one"))
		svi.SetOn(
			"one",
			NewNamespacedKey(system, "one"),
			0,    // we're effectively at genesis right now
			1000, // plan to give this variable to the sys var on 1000
		)

		// simple case: associate a string with a namespaced key
		svi.set("two", NewNamespacedKey(system, "two"))

		// demonstrate that aliasing is possible: the official system name may not
		// be the same as the actual key name
		svi.set("foo", NewNamespacedKey(bpc, "bar"))
	}

	// set the ReleaseFromEndowmentsKeys indirect to a BPC variable
	svi.set(
		sv.ReleaseFromEndowmentAddressName,
		NewNamespacedKey(bpc, sv.ReleaseFromEndowmentAddressName),
	)

	// set the rate table indirects to a bpc variable
	svi.set(
		sv.UnlockedRateTableName,
		NewNamespacedKey(bpc, sv.UnlockedRateTableName),
	)
	svi.set(sv.LockedRateTableName,
		NewNamespacedKey(bpc, sv.LockedRateTableName),
	)

	svi.set(
		sv.DefaultSettlementDurationName,
		NewNamespacedKey(bpc, sv.DefaultSettlementDurationName),
	)

	svi.set(
		sv.TxFeeScriptName,
		NewNamespacedKey(bpc, sv.TxFeeScriptName),
	)

	svi.set(
		sv.MinStakeName,
		NewNamespacedKey(bpc, sv.MinStakeName),
	)

	svi.set(
		sv.NodeGoodnessFuncName,
		NewNamespacedKey(bpc, sv.NodeGoodnessFuncName),
	)

	svi.set(
		sv.EAIFeeTableName,
		NewNamespacedKey(bpc, sv.EAIFeeTableName),
	)

	svi.set(
		sv.MinDurationBetweenNodeRewardNominationsName,
		NewNamespacedKey(bpc, sv.MinDurationBetweenNodeRewardNominationsName),
	)

	svi.set(
		sv.NominateNodeRewardAddressName,
		NewNamespacedKey(bpc, sv.NominateNodeRewardAddressName),
	)

	svi.set(
		sv.NodeRewardNominationTimeoutName,
		NewNamespacedKey(bpc, sv.NodeRewardNominationTimeoutName),
	)

	return svi
}

func makeMockEAIFeeTable() sv.EAIFeeTable {
	return sv.EAIFeeTable{
		makeMockEAIFee("ndev operations", 40),
		makeMockEAIFee("ntrd operations", 10),
		makeMockEAIFee("rfe account", 1),
		makeMockEAIFee("rewards nomination acct", 1),
		makeMockNodeRewardEAIFee(98),
	}
}

func makeMockEAIFee(_ string, thousandths int64) sv.EAIFee {
	public, _, err := signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
	}
	addr, err := address.Generate(address.KindNdau, public.KeyBytes())
	if err != nil {
		panic(err)
	}
	fee, err := signed.MulDiv(thousandths, constants.QuantaPerUnit, 1000)
	if err != nil {
		panic(err)
	}
	return sv.EAIFee{
		Fee: math.Ndau(fee),
		To:  &addr,
	}
}

func makeMockNodeRewardEAIFee(thousandths int64) sv.EAIFee {
	fee, err := signed.MulDiv(thousandths, constants.QuantaPerUnit, 1000)
	if err != nil {
		panic(err)
	}
	return sv.EAIFee{
		Fee: math.Ndau(fee),
		To:  nil,
	}
}
