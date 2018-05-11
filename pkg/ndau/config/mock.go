package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type kv struct {
	Key   B64Data
	Value B64Data
}

type namespace struct {
	Namespace B64Data
	Data      []kv
}

type cm struct {
	Namespaces []namespace
}

// ChaosMock is a mocked representation of the chaos chain.
//
// (De)serialization is a pain: despite the TOML format supporting
// map[B64Data]map[B64Data]B64data, the go language prohibits that.
// Instead, we're stuck with lists of structs.
type ChaosMock map[string]map[string][]byte
type chaosMockInner map[string][]byte

// Static type assertion that ChaosMock implements the SystemStore interface
var _ SystemStore = (*ChaosMock)(nil)

// DefaultMockPath returns the default path at which a chaos mock file is expected
func DefaultMockPath(ndauhome string) string {
	return filepath.Join(ndauhome, "ndau", "mock-chaos.toml")
}

// LoadMock returns a ChaosMock loaded from its file
func LoadMock(mockPath string) (ChaosMock, error) {
	cmRaw := new(cm)
	_, err := toml.DecodeFile(mockPath, cmRaw)
	if err != nil {
		return nil, err
	}

	mock := make(ChaosMock)
	for _, ns := range cmRaw.Namespaces {
		name := string(ns.Namespace.Bytes())
		mock[name] = make(chaosMockInner)
		for _, kvi := range ns.Data {
			k := string(kvi.Key.Bytes())
			mock[name][k] = kvi.Value.Bytes()
		}
	}
	return mock, nil
}

// Dump stores the given ChaosMock in a file
func (m ChaosMock) Dump(mockPath string) error {
	namespaces := make([]namespace, 0)
	for ns, mData := range m {
		data := make([]kv, 0)
		for k, v := range mData {
			data = append(data, kv{
				Key:   NewB64Data([]byte(k)),
				Value: NewB64Data([]byte(v)),
			})
		}
		namespaces = append(namespaces, namespace{
			Namespace: NewB64Data([]byte(ns)),
			Data:      data,
		})
	}
	chaosMock := cm{
		Namespaces: namespaces,
	}

	fp, err := os.Create(mockPath)
	defer fp.Close()
	if err != nil {
		return err
	}
	return toml.NewEncoder(fp).Encode(chaosMock)
}

// Set puts val into the mock ns and key
func (m ChaosMock) Set(ns, key string, val []byte) {
	if m == nil {
		m = make(ChaosMock)
	}
	if _, ok := m[ns]; !ok {
		m[ns] = make(chaosMockInner)
	}
	m[ns][key] = val
}

// Sets puts the string val into the mock ns and key
func (m ChaosMock) Sets(ns, key, val string) {
	m.Set(ns, key, []byte(val))
}

// Get implements the SystemStore interface
func (m ChaosMock) Get(namespace, key []byte) ([]byte, error) {
	inner, hasNamespace := m[string(namespace)]
	if hasNamespace {
		value, hasKey := inner[string(key)]
		if hasKey {
			return value, nil
		}
		return nil, errors.New("Requested key does not exist")
	}
	return nil, errors.New("Requested namespace does not exist")
}
