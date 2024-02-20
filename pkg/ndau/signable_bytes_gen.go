package ndau

// This file generated by txgen: https://github.com/ndau/generator/pkg/txgen
// DO NOT EDIT

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	"github.com/ndau/metanode/pkg/meta/transaction"
)

func intbytes(i int64) []byte {
	ib := make([]byte, 8)
	binary.BigEndian.PutUint64(ib, uint64(i))
	return ib
}

func bytesOf(field interface{}) []byte {
	switch x := field.(type) {
	case json.Number:
		i, err := x.Int64()
		if err != nil {
			return nil
		}
		return intbytes(i)
	case string:
		return []byte(x)
	case bool:
		var b []byte
		if x {
			b = []byte{0x01}
		} else {
			b = []byte{0x00}
		}
		return b
	}
	// for lists and maps, we have no choice but to go reflective:
	// https://stackoverflow.com/a/38748189/504550
	v := reflect.ValueOf(field)
	switch v.Kind() {
	case reflect.Invalid:
		return []byte{}
	case reflect.Slice:
		out := make([]byte, 0)
		for idx := 0; idx < v.Len(); idx++ {
			out = append(out, bytesOf(v.Index(idx).Interface())...)
		}
		return out
	case reflect.Map:
		// first, get the keys as a list
		keys := make([]reflect.Value, 0, v.Len())
		for _, k := range v.MapKeys() {
			if k.Kind() != reflect.String {
				panic("json dict had non-string key: " + k.Kind().String() + ": " + k.String())
			}
			s := k.Interface().(string)
			if !strings.EqualFold(s, "signature") && !strings.EqualFold(s, "signatures") {
				keys = append(keys, k)
			}
		}
		// sort the keys list for a stable iteration order
		sort.Slice(keys, func(i, j int) bool { return keys[i].Interface().(string) < keys[j].Interface().(string) })

		// for each field in the sorted set, append the bytes of that field to the output
		out := make([]byte, 0)
		for _, key := range keys {
			b := bytesOf(v.MapIndex(key).Interface())
			if b == nil {
				return nil
			}
			out = append(out, b...)
		}
		return out
	}

	panic("unknown field type: " + v.Kind().String())
}

// For byte-compatible interoperability with other languages, JSON in particular,
// we can't simply operate on the bytes and native methods of our transaction.
// Instead, we must first convert the transaction to JSON, and then operate on
// the types available to JSON: strings, numbers, bools, lists, and maps.
func sbOf(tx metatx.Transactable) []byte {
	jstext, err := json.Marshal(tx)
	if err != nil {
		return nil
	}

	buffer := bytes.NewBuffer(jstext)
	jsmap := make(map[string]interface{})

	jsdec := json.NewDecoder(buffer)
	jsdec.UseNumber()
	err = jsdec.Decode(&jsmap)
	if err != nil {
		return nil
	}

	return bytesOf(jsmap)
}

// SignableBytes partially implements metatx.Transactable for Transfer
func (tx *Transfer) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for ChangeValidation
func (tx *ChangeValidation) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for ReleaseFromEndowment
func (tx *ReleaseFromEndowment) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for ChangeRecoursePeriod
func (tx *ChangeRecoursePeriod) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for Delegate
func (tx *Delegate) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for CreditEAI
func (tx *CreditEAI) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for Lock
func (tx *Lock) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for Notify
func (tx *Notify) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for SetRewardsDestination
func (tx *SetRewardsDestination) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for SetValidation
func (tx *SetValidation) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for Stake
func (tx *Stake) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for RegisterNode
func (tx *RegisterNode) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for NominateNodeReward
func (tx *NominateNodeReward) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for ClaimNodeReward
func (tx *ClaimNodeReward) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for TransferAndLock
func (tx *TransferAndLock) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for CommandValidatorChange
func (tx *CommandValidatorChange) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for UnregisterNode
func (tx *UnregisterNode) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for Unstake
func (tx *Unstake) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for Issue
func (tx *Issue) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for CreateChildAccount
func (tx *CreateChildAccount) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for RecordPrice
func (tx *RecordPrice) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for SetSysvar
func (tx *SetSysvar) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for SetStakeRules
func (tx *SetStakeRules) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for RecordEndowmentNAV
func (tx *RecordEndowmentNAV) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for ResolveStake
func (tx *ResolveStake) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for Burn
func (tx *Burn) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for Burn
func (tx *Burn) SignableBytes() []byte {
	return sbOf(tx)
}

// SignableBytes partially implements metatx.Transactable for ChangeSchema
func (tx *ChangeSchema) SignableBytes() []byte {
	return sbOf(tx)
}


