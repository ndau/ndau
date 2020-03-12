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
	"sort"
	"strings"

	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/ndau/ndaumath/pkg/signature"
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

var nodeFieldNames []string
var nodeStructTemplate nt.StructTemplate

func init() {
	initNodeStructTemplate(nil)
}

func initNodeStructTemplate(managedFields []string) {
	nodeFieldNames = []string{
		"Active",
		"DistributionScript",
		"Key",
		"TMAddress",
	}
	if len(managedFields) > 0 {
		nodeFieldNames = append(nodeFieldNames, managedFields...)
		sort.Sort(sort.StringSlice(nodeFieldNames))
	}
	nodeStructTemplate = nt.MakeStructTemplate("Node", nodeFieldNames)
}

func needNodeStructTemplateInit(managedFields []string) bool {
	// Loop over the full field name list and make sure that every managed var in it also appears
	// in the given managed field list.  They are both sorted, so the loop is O(linear).
	i := 0
	iLimit := len(managedFields)
	for _, fieldName := range nodeFieldNames {
		if strings.HasPrefix(fieldName, "managedVar") || strings.HasPrefix(fieldName, "HasmanagedVar") {
			if i == iLimit || managedFields[i] != fieldName {
				// We found a managed var in the full list that wasn't in the given list,
				// or the managed field name in sorted order doesn't match; re-init.
				return true
			}
			i++
			// Keep going even if i == iLimit, to ensure no other managed vars in the full list.
		}
	}

	// Re-init if we didn't find all of the given managed fields in the full list.
	return i != iLimit
}

// IsManagedVarSet returns whether the given managed var has ever been set in the Node.
func (x *Node) IsManagedVarSet(name string) bool {
	if x.managedVars == nil {
		return false
	}
	_, ok := x.managedVars[name]
	return ok
}

// Ensure the managed vars map exists and has the given name set as one of its keys.
func (x *Node) ensureManagedVar(name string) {
	if x.managedVars == nil {
		x.managedVars = make(map[string]struct{})
	}
	if _, ok := x.managedVars[name]; !ok {
		x.managedVars[name] = struct{}{}
	}
}

// GetRegistration returns the Node struct's managedVarRegistration value.
func (x *Node) GetRegistration() math.Timestamp {
	return x.managedVarRegistration
}

// SetRegistration sets the Node struct's managedVarRegistration value,
// and flags it for noms marshaling if this is the first time it's being set.
func (x *Node) SetRegistration(val math.Timestamp) {
	x.ensureManagedVar("Registration")
	x.managedVarRegistration = val
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x Node) MarshalNoms(vrw nt.ValueReadWriter) (nodeValue nt.Value, err error) {
	// x.Active (bool->*ast.Ident) is primitive: true

	// x.DistributionScript ([]byte->*ast.ArrayType) is primitive: true

	// x.TMAddress (string->*ast.Ident) is primitive: true

	// x.Key (signature.PublicKey->*ast.SelectorExpr) is primitive: false
	// template decompose: x.Key (signature.PublicKey->*ast.SelectorExpr)
	// template textmarshaler: x.Key
	keyString, err := x.Key.MarshalText()
	if err != nil {
		return nil, errors.Wrap(err, "Node.MarshalNoms->Key.MarshalText")
	}

	// x.managedVars (map[string]struct{}->*ast.MapType) is primitive: false
	// template decompose: x.managedVars (map[string]struct{}->*ast.MapType)
	// template set:  x.managedVars
	managedVarsItems := make([]nt.Value, 0, len(x.managedVars))
	if len(x.managedVars) > 0 {
		// We need to iterate the set in sorted order, so build []string and sort it first
		managedVarsSorted := make([]string, 0, len(x.managedVars))
		for managedVarsItem := range x.managedVars {
			managedVarsSorted = append(managedVarsSorted, managedVarsItem)
		}
		sort.Sort(sort.StringSlice(managedVarsSorted))
		for _, managedVarsItem := range managedVarsSorted {
			managedVarsItems = append(
				managedVarsItems,
				nt.String(managedVarsItem),
			)
		}
	}

	// x.managedVarRegistration (math.Timestamp->*ast.SelectorExpr) is primitive: true

	var managedFields []string

	values := make([]nt.Value, 0, 6)
	// x.Active (bool)
	values = append(values, nt.Bool(x.Active))
	// x.DistributionScript ([]byte)
	values = append(values, nt.String(x.DistributionScript))
	// x.Key (signature.PublicKey)
	values = append(values, nt.String(keyString))
	// x.TMAddress (string)
	values = append(values, nt.String(x.TMAddress))
	// x.managedVarRegistration (math.Timestamp)
	if x.IsManagedVarSet("Registration") {
		managedFields = append(managedFields, "managedVarRegistration")
		values = append(values, util.Int(x.managedVarRegistration).NomsValue())
	}
	// x.managedVars (map[string]struct{})
	if x.managedVars != nil {
		managedFields = append(managedFields, "managedVars")
		values = append(values, nt.NewSet(vrw, managedVarsItems...))
	}

	if needNodeStructTemplateInit(managedFields) {
		initNodeStructTemplate(managedFields)
	}

	return nodeStructTemplate.NewStruct(values), nil
}

var _ marshal.Marshaler = (*Node)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *Node) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"Node.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) (stop bool) {
		switch name {
		// x.Active (bool->*ast.Ident) is primitive: true
		case "Active":
			// template u_decompose: x.Active (bool->*ast.Ident)
			// template u_primitive: x.Active
			activeValue, ok := value.(nt.Bool)
			if !ok {
				err = fmt.Errorf(
					"Node.UnmarshalNoms expected value to be a nt.Bool; found %s",
					reflect.TypeOf(value),
				)
			}
			activeTyped := bool(activeValue)

			x.Active = activeTyped
		// x.DistributionScript ([]byte->*ast.ArrayType) is primitive: true
		case "DistributionScript":
			// template u_decompose: x.DistributionScript ([]byte->*ast.ArrayType)
			// template u_primitive: x.DistributionScript
			distributionScriptValue, ok := value.(nt.String)
			if !ok {
				err = fmt.Errorf(
					"Node.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.TypeOf(value),
				)
			}
			distributionScriptTyped := []byte(distributionScriptValue)

			x.DistributionScript = distributionScriptTyped
		// x.TMAddress (string->*ast.Ident) is primitive: true
		case "TMAddress":
			// template u_decompose: x.TMAddress (string->*ast.Ident)
			// template u_primitive: x.TMAddress
			tMAddressValue, ok := value.(nt.String)
			if !ok {
				err = fmt.Errorf(
					"Node.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.TypeOf(value),
				)
			}
			tMAddressTyped := string(tMAddressValue)

			x.TMAddress = tMAddressTyped
		// x.Key (signature.PublicKey->*ast.SelectorExpr) is primitive: false
		case "Key":
			// template u_decompose: x.Key (signature.PublicKey->*ast.SelectorExpr)
			// template u_textmarshaler: x.Key
			var keyValue signature.PublicKey
			if keyString, ok := value.(nt.String); ok {
				err = keyValue.UnmarshalText([]byte(keyString))
			} else {
				err = fmt.Errorf(
					"Node.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.ValueOf(value).Type(),
				)
			}

			x.Key = keyValue
		// x.managedVars (map[string]struct{}->*ast.MapType) is primitive: false
		case "managedVars":
			// template u_decompose: x.managedVars (map[string]struct{}->*ast.MapType)
			// template u_set: x.managedVars
			managedVarsGoSet := make(map[string]struct{})
			if managedVarsSet, ok := value.(nt.Set); ok {
				managedVarsSet.Iter(func(managedVarsItem nt.Value) (stop bool) {
					if managedVarsItemString, ok := managedVarsItem.(nt.String); ok {
						managedVarsGoSet[string(managedVarsItemString)] = struct{}{}
					} else {
						err = fmt.Errorf(
							"Node.AccountData.UnmarshalNoms expected managedVarsItem to be a nt.String; found %s",
							reflect.TypeOf(value),
						)
					}
					return err != nil
				})
			} else {
				err = fmt.Errorf(
					"Node.AccountData.UnmarshalNoms expected managedVars to be a nt.Set; found %s",
					reflect.TypeOf(value),
				)
			}

			x.managedVars = managedVarsGoSet
		// x.managedVarRegistration (math.Timestamp->*ast.SelectorExpr) is primitive: true
		case "managedVarRegistration":
			// template u_decompose: x.managedVarRegistration (math.Timestamp->*ast.SelectorExpr)
			// template u_primitive: x.managedVarRegistration
			var managedVarRegistrationValue util.Int
			managedVarRegistrationValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "Node.UnmarshalNoms->managedVarRegistration")
				return
			}
			managedVarRegistrationTyped := math.Timestamp(managedVarRegistrationValue)

			x.managedVarRegistration = managedVarRegistrationTyped
		}
		stop = err != nil
		return
	})
	return
}

var _ marshal.Unmarshaler = (*Node)(nil)
