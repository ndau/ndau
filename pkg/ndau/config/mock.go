package config

import (
	"os"
	"path/filepath"

	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// ChaosMock is a mocked representation of the chaos chain.
//
// (De)serialization is a pain: despite its storage in binary format,
// and the natural representation of the chain as
// map[wkt.Bytes]map[wkt.Bytes]B64data, the go language prohibits that.
// Instead, we're stuck with lists of structs.
type ChaosMock map[string]map[string][]byte
type chaosMockInner map[string][]byte

// Static type assertion that ChaosMock implements the SystemStore interface
var _ SystemStore = (*ChaosMock)(nil)

// DefaultMockPath returns the default path at which a chaos mock file is expected
func DefaultMockPath(ndauhome string) string {
	return filepath.Join(ndauhome, "ndau", "mock-chaos.msgp")
}

// LoadMock returns a ChaosMock loaded from its file
func LoadMock(mockPath string) (ChaosMock, error) {
	mcc := new(MockChaosChain)
	mockFp, err := os.Open(mockPath)
	if err != nil {
		return nil, err
	}
	defer mockFp.Close()
	bufMockReader := msgp.NewReader(mockFp)
	err = mcc.DecodeMsg(bufMockReader)
	if err != nil {
		return nil, err
	}

	mock := make(ChaosMock)
	for _, ns := range mcc.Namespaces {
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
	namespaces := make([]MockNamespace, 0)
	for ns, mData := range m {
		data := make([]MockKeyValue, 0)
		for k, v := range mData {
			data = append(data, MockKeyValue{
				Key:   wkt.Bytes([]byte(k)),
				Value: wkt.Bytes([]byte(v)),
			})
		}
		namespaces = append(namespaces, MockNamespace{
			Namespace: wkt.Bytes([]byte(ns)),
			Data:      data,
		})
	}
	chaosMock := MockChaosChain{
		Namespaces: namespaces,
	}

	fp, err := os.Create(mockPath)
	if err != nil {
		return err
	}
	defer fp.Close()

	mockWriter := msgp.NewWriter(fp)
	defer mockWriter.Flush()
	return chaosMock.EncodeMsg(mockWriter)
}

// Set puts val into the mock ns and key
func (m ChaosMock) Set(ns []byte, key, val msgp.Marshaler) {
	if m == nil {
		m = make(ChaosMock)
	}
	if _, ok := m[string(ns)]; !ok {
		m[string(ns)] = make(chaosMockInner)
	}
	keyBytes, err := key.MarshalMsg([]byte{})
	if err != nil {
		panic(err)
	}
	valBytes, err := val.MarshalMsg([]byte{})
	if err != nil {
		panic(err)
	}
	m[string(ns)][string(keyBytes)] = valBytes
}

// Sets puts val into the mock ns and key.
//
// key is a string which gets reinterpreted as bytes.
func (m ChaosMock) Sets(ns []byte, key string, val msgp.Marshaler) {
	m.Set(ns, wkt.Bytes([]byte(key)), val)
}

// Get implements the SystemStore interface
func (m ChaosMock) Get(namespace []byte, key msgp.Marshaler, value msgp.Unmarshaler) error {
	inner, hasNamespace := m[string(namespace)]
	if hasNamespace {
		keyBytes, err := key.MarshalMsg([]byte{})
		if err != nil {
			return errors.Wrap(err, "ChaosMock.Get failed to marshal key")
		}
		valueBytes, hasKey := inner[string(keyBytes)]
		if hasKey {
			leftovers, err := value.UnmarshalMsg(valueBytes)
			if len(leftovers) > 0 {
				return errors.New("ChaosMock.Get unmarshal had leftover bytes")
			}
			return errors.Wrap(err, "ChaosMock.Get failed")

		}
		return errors.New("Requested key does not exist")
	}
	return errors.New("Requested namespace does not exist")
}
