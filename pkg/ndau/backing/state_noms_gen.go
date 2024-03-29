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

	"github.com/ndau/noms/go/marshal"
	nt "github.com/ndau/noms/go/types"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/pricecurve"
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

var stateFieldNames []string
var stateStructTemplate nt.StructTemplate

func init() {
	initStateStructTemplate(nil)
}

func initStateStructTemplate(managedFields []string) {
	stateFieldNames = []string{
		"Accounts",
		"Delegates",
		"HasNodeRewardWinner",
		"LastNodeRewardNomination",
		"MarketPrice",
		"NodeRewardWinner",
		"Nodes",
		"PendingNodeReward",
		"SIB",
		"Sysvars",
		"TargetPrice",
		"TotalBurned",
		"TotalIssue",
		"TotalRFE",
		"UnclaimedNodeReward",
	}
	if len(managedFields) > 0 {
		stateFieldNames = append(stateFieldNames, managedFields...)
		sort.Sort(sort.StringSlice(stateFieldNames))
	}
	stateStructTemplate = nt.MakeStructTemplate("State", stateFieldNames)
}

func needStateStructTemplateInit(managedFields []string) bool {
	// Loop over the full field name list and make sure that every managed var in it also appears
	// in the given managed field list.  They are both sorted, so the loop is O(linear).
	i := 0
	iLimit := len(managedFields)
	for _, fieldName := range stateFieldNames {
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

// IsManagedVarSet returns whether the given managed var has ever been set in the State.
func (x *State) IsManagedVarSet(name string) bool {
	if x.managedVars == nil {
		return false
	}
	_, ok := x.managedVars[name]
	return ok
}

// Ensure the managed vars map exists and has the given name set as one of its keys.
func (x *State) ensureManagedVar(name string) {
	if x.managedVars == nil {
		x.managedVars = make(map[string]struct{})
	}
	if _, ok := x.managedVars[name]; !ok {
		x.managedVars[name] = struct{}{}
	}
}

// GetEndowmentNAV returns the State struct's managedVarEndowmentNAV value.
func (x *State) GetEndowmentNAV() pricecurve.Nanocent {
	return x.managedVarEndowmentNAV
}

// SetEndowmentNAV sets the State struct's managedVarEndowmentNAV value,
// and flags it for noms marshaling if this is the first time it's being set.
func (x *State) SetEndowmentNAV(val pricecurve.Nanocent) {
	x.ensureManagedVar("EndowmentNAV")
	x.managedVarEndowmentNAV = val
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x State) MarshalNoms(vrw nt.ValueReadWriter) (stateValue nt.Value, err error) {
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

	// x.Accounts (map[string]AccountData->*ast.MapType) is primitive: false
	// template decompose: x.Accounts (map[string]AccountData->*ast.MapType)
	// template map: x.Accounts
	accountsKVs := make([]nt.Value, 0, len(x.Accounts)*2)
	for accountsKey, accountsValue := range x.Accounts {
		// template decompose: accountsValue (AccountData->*ast.Ident)
		// template nomsmarshaler: accountsValue
		accountsValueValue, err := accountsValue.MarshalNoms(vrw)
		if err != nil {
			return nil, errors.Wrap(err, "State.MarshalNoms->accountsValue.MarshalNoms")
		}
		accountsKVs = append(
			accountsKVs,
			nt.String(accountsKey),
			accountsValueValue,
		)
	}

	// x.Delegates (map[string]map[string]struct{}->*ast.MapType) is primitive: false
	// template decompose: x.Delegates (map[string]map[string]struct{}->*ast.MapType)
	// template map: x.Delegates
	delegatesKVs := make([]nt.Value, 0, len(x.Delegates)*2)
	for delegatesKey, delegatesValue := range x.Delegates {
		// template decompose: delegatesValue (map[string]struct{}->*ast.MapType)
		// template set:  delegatesValue
		delegatesValueItems := make([]nt.Value, 0, len(delegatesValue))
		if len(delegatesValue) > 0 {
			// We need to iterate the set in sorted order, so build []string and sort it first
			delegatesValueSorted := make([]string, 0, len(delegatesValue))
			for delegatesValueItem := range delegatesValue {
				delegatesValueSorted = append(delegatesValueSorted, delegatesValueItem)
			}
			sort.Sort(sort.StringSlice(delegatesValueSorted))
			for _, delegatesValueItem := range delegatesValueSorted {
				delegatesValueItems = append(
					delegatesValueItems,
					nt.String(delegatesValueItem),
				)
			}
		}
		delegatesKVs = append(
			delegatesKVs,
			nt.String(delegatesKey),
			nt.NewSet(vrw, delegatesValueItems...),
		)
	}

	// x.Nodes (map[string]Node->*ast.MapType) is primitive: false
	// template decompose: x.Nodes (map[string]Node->*ast.MapType)
	// template map: x.Nodes
	nodesKVs := make([]nt.Value, 0, len(x.Nodes)*2)
	for nodesKey, nodesValue := range x.Nodes {
		// template decompose: nodesValue (Node->*ast.Ident)
		// template nomsmarshaler: nodesValue
		nodesValueValue, err := nodesValue.MarshalNoms(vrw)
		if err != nil {
			return nil, errors.Wrap(err, "State.MarshalNoms->nodesValue.MarshalNoms")
		}
		nodesKVs = append(
			nodesKVs,
			nt.String(nodesKey),
			nodesValueValue,
		)
	}

	// x.LastNodeRewardNomination (math.Timestamp->*ast.SelectorExpr) is primitive: true

	// x.PendingNodeReward (math.Ndau->*ast.SelectorExpr) is primitive: true

	// x.UnclaimedNodeReward (math.Ndau->*ast.SelectorExpr) is primitive: true

	// x.NodeRewardWinner (*address.Address->*ast.StarExpr) is primitive: false
	// template decompose: x.NodeRewardWinner (*address.Address->*ast.StarExpr)
	// template pointer:  x.NodeRewardWinner
	var nodeRewardWinnerUnptr nt.Value
	if x.NodeRewardWinner == nil {
		nodeRewardWinnerUnptr = nt.String("")
	} else {
		// template decompose: (*x.NodeRewardWinner) (address.Address->*ast.SelectorExpr)
		// template textmarshaler: (*x.NodeRewardWinner)
		nodeRewardWinnerString, err := (*x.NodeRewardWinner).MarshalText()
		if err != nil {
			return nil, errors.Wrap(err, "State.MarshalNoms->NodeRewardWinner.MarshalText")
		}
		nodeRewardWinnerUnptr = nt.String(nodeRewardWinnerString)
	}

	// x.TotalRFE (math.Ndau->*ast.SelectorExpr) is primitive: true

	// x.TotalIssue (math.Ndau->*ast.SelectorExpr) is primitive: true

	// x.SIB (eai.Rate->*ast.SelectorExpr) is primitive: true

	// x.TotalBurned (math.Ndau->*ast.SelectorExpr) is primitive: true

	// x.MarketPrice (pricecurve.Nanocent->*ast.SelectorExpr) is primitive: true

	// x.TargetPrice (pricecurve.Nanocent->*ast.SelectorExpr) is primitive: true

	// x.managedVarEndowmentNAV (pricecurve.Nanocent->*ast.SelectorExpr) is primitive: true

	// x.Sysvars (map[string][]byte->*ast.MapType) is primitive: false
	// template decompose: x.Sysvars (map[string][]byte->*ast.MapType)
	// template map: x.Sysvars
	sysvarsKVs := make([]nt.Value, 0, len(x.Sysvars)*2)
	for sysvarsKey, sysvarsValue := range x.Sysvars {
		// template decompose: sysvarsValue ([]byte->*ast.ArrayType)
		sysvarsKVs = append(
			sysvarsKVs,
			nt.String(sysvarsKey),
			nt.String(sysvarsValue),
		)
	}

	var managedFields []string

	values := make([]nt.Value, 0, 17)
	// x.Accounts (map[string]AccountData)
	values = append(values, nt.NewMap(vrw, accountsKVs...))
	// x.Delegates (map[string]map[string]struct{})
	values = append(values, nt.NewMap(vrw, delegatesKVs...))
	// x.HasNodeRewardWinner (bool)
	values = append(values, nt.Bool(x.NodeRewardWinner != nil))
	// x.LastNodeRewardNomination (math.Timestamp)
	values = append(values, util.Int(x.LastNodeRewardNomination).NomsValue())
	// x.MarketPrice (pricecurve.Nanocent)
	values = append(values, util.Int(x.MarketPrice).NomsValue())
	// x.NodeRewardWinner (*address.Address)
	values = append(values, nodeRewardWinnerUnptr)
	// x.Nodes (map[string]Node)
	values = append(values, nt.NewMap(vrw, nodesKVs...))
	// x.PendingNodeReward (math.Ndau)
	values = append(values, util.Int(x.PendingNodeReward).NomsValue())
	// x.SIB (eai.Rate)
	values = append(values, util.Int(x.SIB).NomsValue())
	// x.Sysvars (map[string][]byte)
	values = append(values, nt.NewMap(vrw, sysvarsKVs...))
	// x.TargetPrice (pricecurve.Nanocent)
	values = append(values, util.Int(x.TargetPrice).NomsValue())
	// x.TotalBurned (math.Ndau)
	values = append(values, util.Int(x.TotalBurned).NomsValue())
	// x.TotalIssue (math.Ndau)
	values = append(values, util.Int(x.TotalIssue).NomsValue())
	// x.TotalRFE (math.Ndau)
	values = append(values, util.Int(x.TotalRFE).NomsValue())
	// x.UnclaimedNodeReward (math.Ndau)
	values = append(values, util.Int(x.UnclaimedNodeReward).NomsValue())
	// x.managedVarEndowmentNAV (pricecurve.Nanocent)
	if x.IsManagedVarSet("EndowmentNAV") {
		managedFields = append(managedFields, "managedVarEndowmentNAV")
		values = append(values, util.Int(x.managedVarEndowmentNAV).NomsValue())
	}
	// x.managedVars (map[string]struct{})
	if x.managedVars != nil {
		managedFields = append(managedFields, "managedVars")
		values = append(values, nt.NewSet(vrw, managedVarsItems...))
	}

	if needStateStructTemplateInit(managedFields) {
		initStateStructTemplate(managedFields)
	}

	return stateStructTemplate.NewStruct(values), nil
}

var _ marshal.Marshaler = (*State)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *State) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"State.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) (stop bool) {
		switch name {
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
							"State.AccountData.UnmarshalNoms expected managedVarsItem to be a nt.String; found %s",
							reflect.TypeOf(value),
						)
					}
					return err != nil
				})
			} else {
				err = fmt.Errorf(
					"State.AccountData.UnmarshalNoms expected managedVars to be a nt.Set; found %s",
					reflect.TypeOf(value),
				)
			}

			x.managedVars = managedVarsGoSet
		// x.Accounts (map[string]AccountData->*ast.MapType) is primitive: false
		case "Accounts":
			// template u_decompose: x.Accounts (map[string]AccountData->*ast.MapType)
			// template u_map: x.Accounts
			accountsGMap := make(map[string]AccountData)
			if accountsNMap, ok := value.(nt.Map); ok {
				accountsNMap.Iter(func(accountsKey, accountsValue nt.Value) (stop bool) {
					accountsKeyString, ok := accountsKey.(nt.String)
					if !ok {
						err = fmt.Errorf(
							"State.UnmarshalNoms expected accountsKey to be a nt.String; found %s",
							reflect.TypeOf(accountsKey),
						)
						return true
					}

					// template u_decompose: accountsValue (AccountData->*ast.Ident)
					// template u_nomsmarshaler: accountsValue
					var accountsValueInstance AccountData
					err = accountsValueInstance.UnmarshalNoms(accountsValue)
					err = errors.Wrap(err, "State.UnmarshalNoms->accountsValue")
					if err != nil {
						return true
					}
					accountsGMap[string(accountsKeyString)] = accountsValueInstance
					return false
				})
			} else {
				err = fmt.Errorf(
					"State.UnmarshalNoms expected accountsGMap to be a nt.Map; found %s",
					reflect.TypeOf(value),
				)
			}

			x.Accounts = accountsGMap
		// x.Delegates (map[string]map[string]struct{}->*ast.MapType) is primitive: false
		case "Delegates":
			// template u_decompose: x.Delegates (map[string]map[string]struct{}->*ast.MapType)
			// template u_map: x.Delegates
			delegatesGMap := make(map[string]map[string]struct{})
			if delegatesNMap, ok := value.(nt.Map); ok {
				delegatesNMap.Iter(func(delegatesKey, delegatesValue nt.Value) (stop bool) {
					delegatesKeyString, ok := delegatesKey.(nt.String)
					if !ok {
						err = fmt.Errorf(
							"State.UnmarshalNoms expected delegatesKey to be a nt.String; found %s",
							reflect.TypeOf(delegatesKey),
						)
						return true
					}

					// template u_decompose: delegatesValue (map[string]struct{}->*ast.MapType)
					// template u_set: delegatesValue
					delegatesValueGoSet := make(map[string]struct{})
					if delegatesValueSet, ok := delegatesValue.(nt.Set); ok {
						delegatesValueSet.Iter(func(delegatesValueItem nt.Value) (stop bool) {
							if delegatesValueItemString, ok := delegatesValueItem.(nt.String); ok {
								delegatesValueGoSet[string(delegatesValueItemString)] = struct{}{}
							} else {
								err = fmt.Errorf(
									"State.AccountData.UnmarshalNoms expected delegatesValueItem to be a nt.String; found %s",
									reflect.TypeOf(delegatesValue),
								)
							}
							return err != nil
						})
					} else {
						err = fmt.Errorf(
							"State.AccountData.UnmarshalNoms expected delegatesValue to be a nt.Set; found %s",
							reflect.TypeOf(delegatesValue),
						)
					}
					if err != nil {
						return true
					}
					delegatesGMap[string(delegatesKeyString)] = delegatesValueGoSet
					return false
				})
			} else {
				err = fmt.Errorf(
					"State.UnmarshalNoms expected delegatesGMap to be a nt.Map; found %s",
					reflect.TypeOf(value),
				)
			}

			x.Delegates = delegatesGMap
		// x.Nodes (map[string]Node->*ast.MapType) is primitive: false
		case "Nodes":
			// template u_decompose: x.Nodes (map[string]Node->*ast.MapType)
			// template u_map: x.Nodes
			nodesGMap := make(map[string]Node)
			if nodesNMap, ok := value.(nt.Map); ok {
				nodesNMap.Iter(func(nodesKey, nodesValue nt.Value) (stop bool) {
					nodesKeyString, ok := nodesKey.(nt.String)
					if !ok {
						err = fmt.Errorf(
							"State.UnmarshalNoms expected nodesKey to be a nt.String; found %s",
							reflect.TypeOf(nodesKey),
						)
						return true
					}

					// template u_decompose: nodesValue (Node->*ast.Ident)
					// template u_nomsmarshaler: nodesValue
					var nodesValueInstance Node
					err = nodesValueInstance.UnmarshalNoms(nodesValue)
					err = errors.Wrap(err, "State.UnmarshalNoms->nodesValue")
					if err != nil {
						return true
					}
					nodesGMap[string(nodesKeyString)] = nodesValueInstance
					return false
				})
			} else {
				err = fmt.Errorf(
					"State.UnmarshalNoms expected nodesGMap to be a nt.Map; found %s",
					reflect.TypeOf(value),
				)
			}

			x.Nodes = nodesGMap
		// x.LastNodeRewardNomination (math.Timestamp->*ast.SelectorExpr) is primitive: true
		case "LastNodeRewardNomination":
			// template u_decompose: x.LastNodeRewardNomination (math.Timestamp->*ast.SelectorExpr)
			// template u_primitive: x.LastNodeRewardNomination
			var lastNodeRewardNominationValue util.Int
			lastNodeRewardNominationValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "State.UnmarshalNoms->LastNodeRewardNomination")
				return
			}
			lastNodeRewardNominationTyped := math.Timestamp(lastNodeRewardNominationValue)

			x.LastNodeRewardNomination = lastNodeRewardNominationTyped
		// x.PendingNodeReward (math.Ndau->*ast.SelectorExpr) is primitive: true
		case "PendingNodeReward":
			// template u_decompose: x.PendingNodeReward (math.Ndau->*ast.SelectorExpr)
			// template u_primitive: x.PendingNodeReward
			var pendingNodeRewardValue util.Int
			pendingNodeRewardValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "State.UnmarshalNoms->PendingNodeReward")
				return
			}
			pendingNodeRewardTyped := math.Ndau(pendingNodeRewardValue)

			x.PendingNodeReward = pendingNodeRewardTyped
		// x.UnclaimedNodeReward (math.Ndau->*ast.SelectorExpr) is primitive: true
		case "UnclaimedNodeReward":
			// template u_decompose: x.UnclaimedNodeReward (math.Ndau->*ast.SelectorExpr)
			// template u_primitive: x.UnclaimedNodeReward
			var unclaimedNodeRewardValue util.Int
			unclaimedNodeRewardValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "State.UnmarshalNoms->UnclaimedNodeReward")
				return
			}
			unclaimedNodeRewardTyped := math.Ndau(unclaimedNodeRewardValue)

			x.UnclaimedNodeReward = unclaimedNodeRewardTyped
		// x.NodeRewardWinner (*address.Address->*ast.StarExpr) is primitive: false
		case "NodeRewardWinner":
			// template u_decompose: x.NodeRewardWinner (*address.Address->*ast.StarExpr)
			// template u_pointer:  x.NodeRewardWinner
			if hasNodeRewardWinnerValue, ok := vs.MaybeGet("HasNodeRewardWinner"); ok {
				if hasNodeRewardWinner, ok := hasNodeRewardWinnerValue.(nt.Bool); ok {
					if !hasNodeRewardWinner {
						return
					}
				} else {
					err = fmt.Errorf(
						"State.UnmarshalNoms expected HasNodeRewardWinner to be a nt.Bool; found %s",
						reflect.TypeOf(hasNodeRewardWinnerValue),
					)
					return
				}
			} else {
				err = fmt.Errorf(
					"State.UnmarshalNoms->NodeRewardWinner is a pointer, so expected a HasNodeRewardWinner field: not found",
				)
				return
			}

			// template u_decompose: x.NodeRewardWinner (address.Address->*ast.SelectorExpr)
			// template u_textmarshaler: x.NodeRewardWinner
			var nodeRewardWinnerValue address.Address
			if nodeRewardWinnerString, ok := value.(nt.String); ok {
				err = nodeRewardWinnerValue.UnmarshalText([]byte(nodeRewardWinnerString))
			} else {
				err = fmt.Errorf(
					"State.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.ValueOf(value).Type(),
				)
			}

			x.NodeRewardWinner = &nodeRewardWinnerValue
		// x.TotalRFE (math.Ndau->*ast.SelectorExpr) is primitive: true
		case "TotalRFE":
			// template u_decompose: x.TotalRFE (math.Ndau->*ast.SelectorExpr)
			// template u_primitive: x.TotalRFE
			var totalRFEValue util.Int
			totalRFEValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "State.UnmarshalNoms->TotalRFE")
				return
			}
			totalRFETyped := math.Ndau(totalRFEValue)

			x.TotalRFE = totalRFETyped
		// x.TotalIssue (math.Ndau->*ast.SelectorExpr) is primitive: true
		case "TotalIssue":
			// template u_decompose: x.TotalIssue (math.Ndau->*ast.SelectorExpr)
			// template u_primitive: x.TotalIssue
			var totalIssueValue util.Int
			totalIssueValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "State.UnmarshalNoms->TotalIssue")
				return
			}
			totalIssueTyped := math.Ndau(totalIssueValue)

			x.TotalIssue = totalIssueTyped
		// x.SIB (eai.Rate->*ast.SelectorExpr) is primitive: true
		case "SIB":
			// template u_decompose: x.SIB (eai.Rate->*ast.SelectorExpr)
			// template u_primitive: x.SIB
			var sIBValue util.Int
			sIBValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "State.UnmarshalNoms->SIB")
				return
			}
			sIBTyped := eai.Rate(sIBValue)

			x.SIB = sIBTyped
		// x.TotalBurned (math.Ndau->*ast.SelectorExpr) is primitive: true
		case "TotalBurned":
			// template u_decompose: x.TotalBurned (math.Ndau->*ast.SelectorExpr)
			// template u_primitive: x.TotalBurned
			var totalBurnedValue util.Int
			totalBurnedValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "State.UnmarshalNoms->TotalBurned")
				return
			}
			totalBurnedTyped := math.Ndau(totalBurnedValue)

			x.TotalBurned = totalBurnedTyped
		// x.MarketPrice (pricecurve.Nanocent->*ast.SelectorExpr) is primitive: true
		case "MarketPrice":
			// template u_decompose: x.MarketPrice (pricecurve.Nanocent->*ast.SelectorExpr)
			// template u_primitive: x.MarketPrice
			var marketPriceValue util.Int
			marketPriceValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "State.UnmarshalNoms->MarketPrice")
				return
			}
			marketPriceTyped := pricecurve.Nanocent(marketPriceValue)

			x.MarketPrice = marketPriceTyped
		// x.TargetPrice (pricecurve.Nanocent->*ast.SelectorExpr) is primitive: true
		case "TargetPrice":
			// template u_decompose: x.TargetPrice (pricecurve.Nanocent->*ast.SelectorExpr)
			// template u_primitive: x.TargetPrice
			var targetPriceValue util.Int
			targetPriceValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "State.UnmarshalNoms->TargetPrice")
				return
			}
			targetPriceTyped := pricecurve.Nanocent(targetPriceValue)

			x.TargetPrice = targetPriceTyped
		// x.managedVarEndowmentNAV (pricecurve.Nanocent->*ast.SelectorExpr) is primitive: true
		case "managedVarEndowmentNAV":
			// template u_decompose: x.managedVarEndowmentNAV (pricecurve.Nanocent->*ast.SelectorExpr)
			// template u_primitive: x.managedVarEndowmentNAV
			var managedVarEndowmentNAVValue util.Int
			managedVarEndowmentNAVValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "State.UnmarshalNoms->managedVarEndowmentNAV")
				return
			}
			managedVarEndowmentNAVTyped := pricecurve.Nanocent(managedVarEndowmentNAVValue)

			x.managedVarEndowmentNAV = managedVarEndowmentNAVTyped
		// x.Sysvars (map[string][]byte->*ast.MapType) is primitive: false
		case "Sysvars":
			// template u_decompose: x.Sysvars (map[string][]byte->*ast.MapType)
			// template u_map: x.Sysvars
			sysvarsGMap := make(map[string][]byte)
			if sysvarsNMap, ok := value.(nt.Map); ok {
				sysvarsNMap.Iter(func(sysvarsKey, sysvarsValue nt.Value) (stop bool) {
					sysvarsKeyString, ok := sysvarsKey.(nt.String)
					if !ok {
						err = fmt.Errorf(
							"State.UnmarshalNoms expected sysvarsKey to be a nt.String; found %s",
							reflect.TypeOf(sysvarsKey),
						)
						return true
					}

					// template u_decompose: sysvarsValue ([]byte->*ast.ArrayType)
					// template u_primitive: sysvarsValue
					sysvarsValueValue, ok := sysvarsValue.(nt.String)
					if !ok {
						err = fmt.Errorf(
							"State.UnmarshalNoms expected sysvarsValue to be a nt.String; found %s",
							reflect.TypeOf(sysvarsValue),
						)
					}
					sysvarsValueTyped := []byte(sysvarsValueValue)
					if err != nil {
						return true
					}
					sysvarsGMap[string(sysvarsKeyString)] = sysvarsValueTyped
					return false
				})
			} else {
				err = fmt.Errorf(
					"State.UnmarshalNoms expected sysvarsGMap to be a nt.Map; found %s",
					reflect.TypeOf(value),
				)
			}

			x.Sysvars = sysvarsGMap
		}
		stop = err != nil
		return
	})
	return
}

var _ marshal.Unmarshaler = (*State)(nil)
