package config

import (
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

type cm []namespace

// ChaosMock is a mocked representation of the chaos chain.
//
// (De)serialization is a pain: despite the TOML format supporting
// map[B64Data]map[B64Data]B64data, the go language prohibits that.
// Instead, we're stuck with lists of structs.
type ChaosMock map[string]map[string][]byte
type chaosMockInner map[string][]byte

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
	for _, ns := range *cmRaw {
		name := string(ns.Namespace.Bytes())
		mock[name] = make(chaosMockInner)
		for _, kvi := range ns.Data {
			k := string(kvi.Key.Bytes())
			mock[name][k] = kvi.Value.Bytes()
		}
	}
	return mock, nil
}

// DumpMock stores the given ChaosMock in a file
func (m ChaosMock) DumpMock(mockPath string) error {
	cmRaw := make([]namespace, 0)
	for ns, mData := range m {
		data := make([]kv, 0)
		for k, v := range mData {
			data = append(data, kv{
				Key:   NewB64Data([]byte(k)),
				Value: NewB64Data([]byte(v)),
			})
		}
		cmRaw = append(cmRaw, namespace{
			Namespace: NewB64Data([]byte(ns)),
			Data:      data,
		})
	}

	fp, err := os.Create(mockPath)
	defer fp.Close()
	if err != nil {
		return err
	}
	return toml.NewEncoder(fp).Encode(cmRaw)
}
