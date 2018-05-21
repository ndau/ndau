package config

import (
	"errors"
	"fmt"
)

//go:generate msgp

// B64Data is a byte slice which can marshal/unmarshal itself as b64
//
// This is useful for a variety of fields: namespaces in particular,
// but also anything else which has a natural binary representation
// but no natural textual representation.
type B64Data struct {
	data []byte
}

//msgp:tuple NamespacedKey

// NamespacedKey is a namespace and key which together identify a unique value on the chaos chain.
type NamespacedKey struct {
	Namespace B64Data
	Key       B64Data
}

// SVIDeferredChange is an indirection struct.
//
// It helps address the coordination problem: in order to prevent forks,
// all nodes must update their system indirects simultaneously. Otherwise,
// nodes processing the same block may disagree on the indirect, and therefore
// the value, of a given system variable.
//
// Current should always be the current value at the time of the update,
// whether or not that value is stored in the existing "Current" or "Future"
// section from the previous update.
//
// ChangeOn should always be at least 1 more than the current height at the
// time of an update, and best practice will be to increase the buffer,
// because there is no guarantee that a particular transaction will make it
// onto the expected block.
type SVIDeferredChange struct {
	Current  NamespacedKey
	Future   NamespacedKey
	ChangeOn uint64
}

// SVIMap is a map of names to deferred changes
//
// Its keys are the string names of system variables.
// Its values are deferred changes. It is a logic error
// to update an SVIMap such that for each updated system variable,
// the updated ChangeOn <= the current height,
// or such that the new value of Current is not equal to the actual
// current value, but it is not possible to actually validate this without
// requiring a custom transaction type for SVIMap updates.
//
// The BPC is encouraged to ensure that it always generates valid SVIMap
// updates, as failure to do so will likely lead to forks.
type SVIMap map[string]SVIDeferredChange

// Marshal this SVIMap to a byte slice
func (m *SVIMap) Marshal() ([]byte, error) {
	return m.MarshalMsg([]byte{})
}

// Unmarshal the byte slice into an SVIMap
func (m *SVIMap) Unmarshal(bytes []byte) error {
	remainder, err := m.UnmarshalMsg(bytes)
	if len(remainder) > 0 {
		return errors.New("Unmarshal produced remainder bytes")
	}
	return err
}

// Get the value of a namespaced key at a specififed height
func (m *SVIMap) Get(name string, height uint64) (nsk NamespacedKey, err error) {
	if m == nil {
		err = errors.New("nil SVIMap")
		return
	}
	deferred, hasKey := map[string]SVIDeferredChange(*m)[name]
	if !hasKey {
		err = fmt.Errorf("Key '%s' not present in SVIMap", name)
		return
	}

	if height >= deferred.ChangeOn {
		nsk = deferred.Future
	} else {
		nsk = deferred.Current
	}

	return
}

// SetOn sets the location of a named system variable to a given namespace and key as of a particular block.
func (m *SVIMap) SetOn(name string, nsk NamespacedKey, current, on uint64) (err error) {
	currentNsk, err := m.Get(name, current)
	if err == nil {
		map[string]SVIDeferredChange(*m)[name] = SVIDeferredChange{
			Current:  currentNsk,
			Future:   nsk,
			ChangeOn: on,
		}
	} else {
		_, hasKey := map[string]SVIDeferredChange(*m)[name]
		if !hasKey {
			// error was probably that the key didn't exist
			err = nil
			map[string]SVIDeferredChange(*m)[name] = SVIDeferredChange{
				Current:  nsk,
				Future:   nsk,
				ChangeOn: on,
			}
		}
	}
	return
}

// shorthand to set a nsk for testing purposes
func (m *SVIMap) set(name string, nsk NamespacedKey) error {
	return m.SetOn(name, nsk, 0, 0)
}

// SystemStore types are stores of system variables.
//
// No restriction is placed on their implementation, so long as they
// can get values from namespaced keys.
type SystemStore interface {
	Get(namespace, key []byte) ([]byte, error)
}

// GetNSK gets the requested namespaced key from any SystemStore
func GetNSK(ss SystemStore, nsk NamespacedKey) (out []byte, err error) {
	out, err = ss.Get(nsk.Namespace.Bytes(), nsk.Key.Bytes())
	return
}

// GetSVI returns the System Variable Indirection map from any SystemStore
func GetSVI(ss SystemStore, nsk NamespacedKey) (*SVIMap, error) {
	svib, err := GetNSK(ss, nsk)
	if err != nil {
		return nil, err
	}
	svi := new(SVIMap)
	err = svi.Unmarshal(svib)
	return svi, nil
}
