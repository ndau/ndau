package address

import (
	"hash/crc32"

	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp

// DataStore is an intermediate type between Address and address.Data
//
// Ideally, it would be private, but the limitations imposed by `msgp`
// code generator require that it be public. Users should not bother with this.
type DataStore struct {
	Key        msgp.Raw `msg:"key"`
	Derivation []byte   `msg:"drv"`
	Crc        uint32   `msg:"crc"`
}

func crc(key []byte, drv []byte) uint32 {
	return crc32.ChecksumIEEE(append(key, drv...))
}

func (d Data) asDataStore() (DataStore, error) {
	key, err := signature.MarshalKey(d.algorithm, d.key)
	if err != nil {
		return DataStore{}, err
	}
	return DataStore{
		Key:        key,
		Derivation: d.derivation,
		Crc:        crc(key, d.derivation),
	}, nil
}

func (ds DataStore) validateCrc() bool {
	return crc(ds.Key, ds.Derivation) == ds.Crc
}

func (ds DataStore) asData() (Data, error) {
	algo, key, err := signature.UnmarshalKey(ds.Key)
	if err != nil {
		return Data{}, err
	}
	return DataFromKeyAndDerivation(key, algo, ds.Derivation)
}
