package config

import (
	"io/ioutil"

	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/tinylib/msgp/msgp"
	"golang.org/x/crypto/ed25519"
)

// we redefine the system namespace here instead of depending on the actual
// value from the chaos chain. It doesn't particularly matter whether we get
// it right for the purposes of this test, and it simplifies the dependency
// management story considerably
var system = []byte("system")

// MakeMock creates a mock file with the specififed data.
//
// If `configPath == ""`, the config file is skipped. Otherwise,
// the config file at that path is created and directed to the
// mock file.
func MakeMock(configPath, mockPath string) (config *Config, err error) {
	bpc, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	svi := makeMockSVI(bpc)
	mock, sviKey := makeMockChaos(bpc, svi)

	// make the mock file
	err = mock.Dump(mockPath)
	if err != nil {
		return nil, err
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

	return config, err
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
func MakeTmpMock(tmpdir string) (config *Config, err error) {
	configFile, err := ioutil.TempFile("", "config")
	if err != nil {
		return nil, err
	}
	mockFile, err := ioutil.TempFile("", "mock")
	if err != nil {
		return nil, err
	}
	return MakeMock(configFile.Name(), mockFile.Name())
}

// mock up some chaos data
func makeMockChaos(bpc []byte, svi msgp.Marshaler) (ChaosMock, *NamespacedKey) {
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
	return mock, &sviKey
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

	return svi
}
