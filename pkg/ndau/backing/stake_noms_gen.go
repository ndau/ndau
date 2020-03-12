package backing

// this code generated by github.com/ndau/generator/cmd/nomsify -- DO NOT EDIT

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"
	"reflect"

	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/ndau/ndaumath/pkg/address"
	math "github.com/ndau/ndaumath/pkg/types"
	util "github.com/ndau/noms-util"
	"github.com/pkg/errors"
)

// Adding new fields to a nomsify-able struct:
//
// Managed vars are useful for adding new fields that are marshaled to noms only after they're
// first set, so that app hashes aren't affected until the new fields are actually needed.
//
// A managed vars map is a hash map whose keys are managed variable names.
// The `managedVars map[string]struct{}` field must be manually declared in the struct.
//
// Declare new fields using the "managedVar" prefix.  e.g. `managedVarSomething SomeType`.
// GetSomething() and SetSomething() are generated for public access to the new field.
//
// Once SetSomething() is called for the first time, typically as a result of processing a new
// transaction that uses it, the managed vars map will contain "Something" as a key and the
// value of managedVarSomething will be stored in noms on the next call to MarshalNoms().
// Until then, all new managedVar fields will retain their "zero" values.

var stakeDataStructTemplate nt.StructTemplate

func init() {
	stakeDataStructTemplate = nt.MakeStructTemplate("StakeData", []string{
		"Point",
		"RulesAcct",
		"StakeTo",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x StakeData) MarshalNoms(vrw nt.ValueReadWriter) (stakeDataValue nt.Value, err error) {
	// x.Point (math.Timestamp->*ast.SelectorExpr) is primitive: true

	// x.RulesAcct (address.Address->*ast.SelectorExpr) is primitive: false
	// template decompose: x.RulesAcct (address.Address->*ast.SelectorExpr)
	// template textmarshaler: x.RulesAcct
	rulesAcctString, err := x.RulesAcct.MarshalText()
	if err != nil {
		return nil, errors.Wrap(err, "StakeData.MarshalNoms->RulesAcct.MarshalText")
	}

	// x.StakeTo (address.Address->*ast.SelectorExpr) is primitive: false
	// template decompose: x.StakeTo (address.Address->*ast.SelectorExpr)
	// template textmarshaler: x.StakeTo
	stakeToString, err := x.StakeTo.MarshalText()
	if err != nil {
		return nil, errors.Wrap(err, "StakeData.MarshalNoms->StakeTo.MarshalText")
	}

	values := []nt.Value{
		// x.Point (math.Timestamp)
		util.Int(x.Point).NomsValue(),
		// x.RulesAcct (address.Address)
		nt.String(rulesAcctString),
		// x.StakeTo (address.Address)
		nt.String(stakeToString),
	}

	return stakeDataStructTemplate.NewStruct(values), nil
}

var _ marshal.Marshaler = (*StakeData)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *StakeData) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"StakeData.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) (stop bool) {
		switch name {
		// x.Point (math.Timestamp->*ast.SelectorExpr) is primitive: true
		case "Point":
			// template u_decompose: x.Point (math.Timestamp->*ast.SelectorExpr)
			// template u_primitive: x.Point
			var pointValue util.Int
			pointValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "StakeData.UnmarshalNoms->Point")
				return
			}
			pointTyped := math.Timestamp(pointValue)

			x.Point = pointTyped
		// x.RulesAcct (address.Address->*ast.SelectorExpr) is primitive: false
		case "RulesAcct":
			// template u_decompose: x.RulesAcct (address.Address->*ast.SelectorExpr)
			// template u_textmarshaler: x.RulesAcct
			var rulesAcctValue address.Address
			if rulesAcctString, ok := value.(nt.String); ok {
				err = rulesAcctValue.UnmarshalText([]byte(rulesAcctString))
			} else {
				err = fmt.Errorf(
					"StakeData.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.ValueOf(value).Type(),
				)
			}

			x.RulesAcct = rulesAcctValue
		// x.StakeTo (address.Address->*ast.SelectorExpr) is primitive: false
		case "StakeTo":
			// template u_decompose: x.StakeTo (address.Address->*ast.SelectorExpr)
			// template u_textmarshaler: x.StakeTo
			var stakeToValue address.Address
			if stakeToString, ok := value.(nt.String); ok {
				err = stakeToValue.UnmarshalText([]byte(stakeToString))
			} else {
				err = fmt.Errorf(
					"StakeData.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.ValueOf(value).Type(),
				)
			}

			x.StakeTo = stakeToValue
		}
		stop = err != nil
		return
	})
	return
}

var _ marshal.Unmarshaler = (*StakeData)(nil)

var stakeRulesStructTemplate nt.StructTemplate

func init() {
	stakeRulesStructTemplate = nt.MakeStructTemplate("StakeRules", []string{
		"Inbound",
		"Script",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x StakeRules) MarshalNoms(vrw nt.ValueReadWriter) (stakeRulesValue nt.Value, err error) {
	// x.Script ([]byte->*ast.ArrayType) is primitive: true

	// x.Inbound (map[string]uint64->*ast.MapType) is primitive: false
	// template decompose: x.Inbound (map[string]uint64->*ast.MapType)
	// template map: x.Inbound
	inboundKVs := make([]nt.Value, 0, len(x.Inbound)*2)
	for inboundKey, inboundValue := range x.Inbound {
		// template decompose: inboundValue (uint64->*ast.Ident)
		inboundKVs = append(
			inboundKVs,
			nt.String(inboundKey),
			util.Int(inboundValue).NomsValue(),
		)
	}

	values := []nt.Value{
		// x.Inbound (map[string]uint64)
		nt.NewMap(vrw, inboundKVs...),
		// x.Script ([]byte)
		nt.String(x.Script),
	}

	return stakeRulesStructTemplate.NewStruct(values), nil
}

var _ marshal.Marshaler = (*StakeRules)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *StakeRules) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"StakeRules.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) (stop bool) {
		switch name {
		// x.Script ([]byte->*ast.ArrayType) is primitive: true
		case "Script":
			// template u_decompose: x.Script ([]byte->*ast.ArrayType)
			// template u_primitive: x.Script
			scriptValue, ok := value.(nt.String)
			if !ok {
				err = fmt.Errorf(
					"StakeRules.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.TypeOf(value),
				)
			}
			scriptTyped := []byte(scriptValue)

			x.Script = scriptTyped
		// x.Inbound (map[string]uint64->*ast.MapType) is primitive: false
		case "Inbound":
			// template u_decompose: x.Inbound (map[string]uint64->*ast.MapType)
			// template u_map: x.Inbound
			inboundGMap := make(map[string]uint64)
			if inboundNMap, ok := value.(nt.Map); ok {
				inboundNMap.Iter(func(inboundKey, inboundValue nt.Value) (stop bool) {
					inboundKeyString, ok := inboundKey.(nt.String)
					if !ok {
						err = fmt.Errorf(
							"StakeRules.UnmarshalNoms expected inboundKey to be a nt.String; found %s",
							reflect.TypeOf(inboundKey),
						)
						return true
					}

					// template u_decompose: inboundValue (uint64->*ast.Ident)
					// template u_primitive: inboundValue
					var inboundValueValue util.Int
					inboundValueValue, err = util.IntFrom(inboundValue)
					if err != nil {
						err = errors.Wrap(err, "StakeRules.UnmarshalNoms->inboundValue")
						return
					}
					inboundValueTyped := uint64(inboundValueValue)
					if err != nil {
						return true
					}
					inboundGMap[string(inboundKeyString)] = inboundValueTyped
					return false
				})
			} else {
				err = fmt.Errorf(
					"StakeRules.UnmarshalNoms expected inboundGMap to be a nt.Map; found %s",
					reflect.TypeOf(value),
				)
			}

			x.Inbound = inboundGMap
		}
		stop = err != nil
		return
	})
	return
}

var _ marshal.Unmarshaler = (*StakeRules)(nil)
