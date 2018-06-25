package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/ndaunode/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/signature/pkg/signature"
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

	svi := makeMockSVI(bpc)
	mock, ma, sviKey := makeMockChaos(bpc, svi)

	// make the mock file
	err = mock.Dump(mockPath)
	if err != nil {
		return nil, nil, err
	}

	if configPath != "" {
		config = &Config{
			ChaosAddress:           "",
			UseMock:                mockPath,
			SystemVariableIndirect: *sviKey,
			ChaosTimeout:           500,
		}
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
func makeMockChaos(bpc []byte, svi msgp.Marshaler) (ChaosMock, MockAssociated, *NamespacedKey) {
	mock := make(ChaosMock)
	mock.Sets(system, "one", wkt.String("system value one"))
	mock.Sets(system, "two", wkt.String("system value two"))

	mock.Sets(bpc, "one", wkt.String("bpc val one"))
	mock.Sets(bpc, "bar", wkt.String("baz"))
	var sviKey NamespacedKey
	if svi != nil {
		mock.Sets(bpc, "svi", svi)
		sviKey = NewNamespacedKey(bpc, "svi")
	}

	// prepare to return associated data
	ma := make(MockAssociated)

	// make RFE keypairs
	const noKeys = 2
	rfeKeys := make(sv.ReleaseFromEndowmentKeys, noKeys)
	rfePrivate := make([]signature.PrivateKey, noKeys)
	var err error
	for i := 0; i < noKeys; i++ {
		rfeKeys[i], rfePrivate[i], err = signature.Generate(signature.Ed25519, nil)
		if err != nil {
			panic(err)
		}
	}
	mock.Sets(bpc, sv.ReleaseFromEndowmentKeysName, rfeKeys)
	ma[sv.ReleaseFromEndowmentKeysName] = rfePrivate

	// make default escrow duration
	ded := sv.DefaultEscrowDuration{math.Day * 15}
	mock.Sets(bpc, sv.DefaultEscrowDurationName, ded)

	return mock, ma, &sviKey
}

// mock up an SVI Map using most of its features
func makeMockSVI(bpc []byte) SVIMap {
	svi := make(SVIMap)
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

	// set the ReleaseFromEndowmentsKeys indirect to a BPC variable
	svi.set(
		sv.ReleaseFromEndowmentKeysName,
		NewNamespacedKey(bpc, sv.ReleaseFromEndowmentKeysName),
	)

	svi.set(
		sv.DefaultEscrowDurationName,
		NewNamespacedKey(bpc, sv.DefaultEscrowDurationName),
	)

	return svi
}