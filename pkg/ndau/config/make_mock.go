package config

import (
	"gitlab.ndau.tech/experiments/chaos-go/pkg/chaos/ns"
	"golang.org/x/crypto/ed25519"
)

// MakeMock creates a mock file with the specififed data.
//
// If `configPath == ""`, the config file is skipped. Otherwise,
// the config file at that path is created and directed to the
// mock file.
func MakeMock(configPath, mockPath string) (err error) {
	bpc, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}

	svi := makeMockSVI(bpc)
	svib, err := svi.Marshal()
	if err != nil {
		return err
	}
	mock, sviKey := makeMockChaos(bpc, svib)

	// make the mock file
	err = mock.Dump(mockPath)
	if err != nil {
		return err
	}

	if configPath != "" {
		config := Config{
			ChaosAddress:           "",
			UseMock:                mockPath,
			SystemVariableIndirect: *sviKey,
		}
		err = config.Dump(configPath)
	}

	return err
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
func makeMockSVI(bpc []byte) *SVIMap {
	svi := new(SVIMap)
	svi.Set("one", NamespacedKey{
		Namespace: NewB64Data(bpc),
		Key:       NewB64Data([]byte("one")),
	})
	svi.SetOn(
		"one",
		NamespacedKey{
			Namespace: NewB64Data(ns.System),
			Key:       NewB64Data([]byte("one")),
		},
		1000, // plan to give this variable to the sys var on 1000
	)

	// simple case: associate a string with a namespaced key
	svi.Set("two", NamespacedKey{
		Namespace: NewB64Data(ns.System),
		Key:       NewB64Data([]byte("two")),
	})

	// demonstrate that aliasing is possible: the official system name may not
	// be the same as the actual key name
	svi.Set("foo", NamespacedKey{
		Namespace: NewB64Data(bpc),
		Key:       NewB64Data([]byte("bar")),
	})

	return svi
}
