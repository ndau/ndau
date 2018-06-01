package address

import (
	"fmt"

	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// Address is the serialized representation of an address.Data struct
type Address []byte

// Data is the data associated with an address
type Data struct {
	key        signature.Key
	algorithm  signature.Algorithm
	derivation []byte
}

// GetKey returns d's key
func (d Data) GetKey() signature.Key {
	return d.key
}

// GetAlgorithm returns d's algorithm
func (d Data) GetAlgorithm() signature.Algorithm {
	return d.algorithm
}

// GetDerivation returns d's derivation
func (d Data) GetDerivation() []byte {
	return d.derivation
}

// DataFromKey constructs a Data object from a key and algorithm
func DataFromKey(key signature.Key, algorithm signature.Algorithm) (Data, error) {
	return DataFromKeyAndDerivation(key, algorithm, nil)
}

// DataFromKeyAndDerivation constructs a Data object from key, algorithm, and derivation
//
// Derivation is useful for HD wallets, but is treated internally as an opaque
// binary blob.
func DataFromKeyAndDerivation(key signature.Key, algorithm signature.Algorithm, derivation []byte) (Data, error) {
	if key == nil {
		return Data{}, fmt.Errorf("nil key")
	}
	if algorithm == nil {
		return Data{}, fmt.Errorf("nil algorithm")
	}
	if len(key) != algorithm.PublicKeySize() {
		return Data{}, fmt.Errorf("Wrong size public key: have %d, need %d", len(key), algorithm.PublicKeySize())
	}
	return Data{
		key:        key,
		algorithm:  algorithm,
		derivation: derivation,
	}, nil
}

// MarshalData serializes address.Data into an Address
func MarshalData(data Data) (Address, error) {
	ds, err := data.asDataStore()
	if err != nil {
		return nil, errors.Wrap(err, "MarshallData fail")
	}
	bytes, err := ds.MarshalMsg(nil)
	err = errors.Wrap(err, "MarshalData fail")
	return bytes, err
}

// UnmarshalData recovers an address.Data from an Address
func UnmarshalData(bytes Address) (Data, error) {
	ds := DataStore{}
	leftovers, err := ds.UnmarshalMsg(bytes)
	if len(leftovers) > 0 {
		err = fmt.Errorf(
			"Unmarshalling provided bytes resulted in %d leftovers",
			len(leftovers),
		)
	}
	if err != nil {
		return Data{}, errors.Wrap(err, "UnmarshalData fail")
	}
	if !ds.validateCrc() {
		return Data{}, errors.New("UnmarshalData fail: DataStore CRC validation failed")
	}
	data, err := ds.asData()
	err = errors.Wrap(err, "UnmarshalData fail")
	return data, err
}
