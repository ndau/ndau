package config

import (
	"io/ioutil"

	"gitlab.ndau.tech/experiments/chaos-go/pkg/chaos/ns"
	"golang.org/x/crypto/ed25519"
)

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
	svib, err := svi.Marshal()
	if err != nil {
		return nil, err
	}
	mock, sviKey := makeMockChaos(bpc, svib)

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
func makeMockChaos(bpc []byte, svib []byte) (ChaosMock, *NamespacedKey) {
	mock := make(ChaosMock)
	sys := string(ns.System)
	mock.Sets(sys, "one", "system value one")
	mock.Sets(sys, "two", "system value two")

	bpcs := string(bpc)
	mock.Sets(bpcs, "one", "bpc val one")
	mock.Sets(bpcs, "bar", "baz")
	var sviKey *NamespacedKey
	if svib != nil {
		mock.Set(bpcs, "svi", svib)
		sviKey = &NamespacedKey{
			Namespace: NewB64Data(bpc),
			Key:       NewB64Data([]byte("svi")),
		}
	}
	return mock, sviKey
}

// mock up an SVI Map using most of its features
func makeMockSVI(bpc []byte) SVIMap {
	svi := make(SVIMap)
	svi.set("one", NamespacedKey{
		Namespace: NewB64Data(bpc),
		Key:       NewB64Data([]byte("one")),
	})
	svi.SetOn(
		"one",
		NamespacedKey{
			Namespace: NewB64Data(ns.System),
			Key:       NewB64Data([]byte("one")),
		},
		0,    // we're effectively at genesis right now
		1000, // plan to give this variable to the sys var on 1000
	)

	// simple case: associate a string with a namespaced key
	svi.set("two", NamespacedKey{
		Namespace: NewB64Data(ns.System),
		Key:       NewB64Data([]byte("two")),
	})

	// demonstrate that aliasing is possible: the official system name may not
	// be the same as the actual key name
	svi.set("foo", NamespacedKey{
		Namespace: NewB64Data(bpc),
		Key:       NewB64Data([]byte("bar")),
	})

	return svi
}
