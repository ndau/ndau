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

	"github.com/ndau/noms/go/marshal"
	nt "github.com/ndau/noms/go/types"
	"github.com/ndau/ndaumath/pkg/address"
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

var accountDataStructTemplate nt.StructTemplate

func init() {
	accountDataStructTemplate = nt.MakeStructTemplate("AccountData", []string{
		"Balance",
		"Costakers",
		"CurrencySeatDate",
		"DelegationNode",
		"HasCurrencySeatDate",
		"HasDelegationNode",
		"HasLock",
		"HasParent",
		"HasProgenitor",
		"HasRewardsTarget",
		"HasStakeRules",
		"Holds",
		"IncomingRewardsFrom",
		"LastEAIUpdate",
		"LastWAAUpdate",
		"Lock",
		"Parent",
		"Progenitor",
		"RecourseSettings",
		"RewardsTarget",
		"Sequence",
		"StakeRules",
		"UncreditedEAI",
		"ValidationKeys",
		"ValidationScript",
		"WeightedAverageAge",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x AccountData) MarshalNoms(vrw nt.ValueReadWriter) (accountDataValue nt.Value, err error) {
	// x.Balance (math.Ndau->*ast.SelectorExpr) is primitive: true

	// x.ValidationKeys ([]signature.PublicKey->*ast.ArrayType) is primitive: false
	// template decompose: x.ValidationKeys ([]signature.PublicKey->*ast.ArrayType)
	// template slice: x.ValidationKeys
	validationKeysItems := make([]nt.Value, 0, len(x.ValidationKeys))
	for _, validationKeysItem := range x.ValidationKeys {
		// template decompose: validationKeysItem (signature.PublicKey->*ast.SelectorExpr)
		// template textmarshaler: validationKeysItem
		validationKeysItemString, err := validationKeysItem.MarshalText()
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->validationKeysItem.MarshalText")
		}
		validationKeysItems = append(
			validationKeysItems,
			nt.String(validationKeysItemString),
		)
	}

	// x.ValidationScript ([]byte->*ast.ArrayType) is primitive: true

	// x.RewardsTarget (*address.Address->*ast.StarExpr) is primitive: false
	// template decompose: x.RewardsTarget (*address.Address->*ast.StarExpr)
	// template pointer:  x.RewardsTarget
	var rewardsTargetUnptr nt.Value
	if x.RewardsTarget == nil {
		rewardsTargetUnptr = nt.String("")
	} else {
		// template decompose: (*x.RewardsTarget) (address.Address->*ast.SelectorExpr)
		// template textmarshaler: (*x.RewardsTarget)
		rewardsTargetString, err := (*x.RewardsTarget).MarshalText()
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->RewardsTarget.MarshalText")
		}
		rewardsTargetUnptr = nt.String(rewardsTargetString)
	}

	// x.IncomingRewardsFrom ([]address.Address->*ast.ArrayType) is primitive: false
	// template decompose: x.IncomingRewardsFrom ([]address.Address->*ast.ArrayType)
	// template slice: x.IncomingRewardsFrom
	incomingRewardsFromItems := make([]nt.Value, 0, len(x.IncomingRewardsFrom))
	for _, incomingRewardsFromItem := range x.IncomingRewardsFrom {
		// template decompose: incomingRewardsFromItem (address.Address->*ast.SelectorExpr)
		// template textmarshaler: incomingRewardsFromItem
		incomingRewardsFromItemString, err := incomingRewardsFromItem.MarshalText()
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->incomingRewardsFromItem.MarshalText")
		}
		incomingRewardsFromItems = append(
			incomingRewardsFromItems,
			nt.String(incomingRewardsFromItemString),
		)
	}

	// x.DelegationNode (*address.Address->*ast.StarExpr) is primitive: false
	// template decompose: x.DelegationNode (*address.Address->*ast.StarExpr)
	// template pointer:  x.DelegationNode
	var delegationNodeUnptr nt.Value
	if x.DelegationNode == nil {
		delegationNodeUnptr = nt.String("")
	} else {
		// template decompose: (*x.DelegationNode) (address.Address->*ast.SelectorExpr)
		// template textmarshaler: (*x.DelegationNode)
		delegationNodeString, err := (*x.DelegationNode).MarshalText()
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->DelegationNode.MarshalText")
		}
		delegationNodeUnptr = nt.String(delegationNodeString)
	}

	// x.Lock (*Lock->*ast.StarExpr) is primitive: false
	// template decompose: x.Lock (*Lock->*ast.StarExpr)
	// template pointer:  x.Lock
	var lockUnptr nt.Value
	if x.Lock == nil {
		lockUnptr = nt.Bool(false)
	} else {
		// template decompose: (*x.Lock) (Lock->*ast.Ident)
		// template nomsmarshaler: (*x.Lock)
		lockValue, err := (*x.Lock).MarshalNoms(vrw)
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->Lock.MarshalNoms")
		}
		lockUnptr = lockValue
	}

	// x.LastEAIUpdate (math.Timestamp->*ast.SelectorExpr) is primitive: true

	// x.LastWAAUpdate (math.Timestamp->*ast.SelectorExpr) is primitive: true

	// x.WeightedAverageAge (math.Duration->*ast.SelectorExpr) is primitive: true

	// x.Sequence (uint64->*ast.Ident) is primitive: true

	// x.StakeRules (*StakeRules->*ast.StarExpr) is primitive: false
	// template decompose: x.StakeRules (*StakeRules->*ast.StarExpr)
	// template pointer:  x.StakeRules
	var stakeRulesUnptr nt.Value
	if x.StakeRules == nil {
		stakeRulesUnptr = nt.Bool(false)
	} else {
		// template decompose: (*x.StakeRules) (StakeRules->*ast.Ident)
		// template nomsmarshaler: (*x.StakeRules)
		stakeRulesValue, err := (*x.StakeRules).MarshalNoms(vrw)
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->StakeRules.MarshalNoms")
		}
		stakeRulesUnptr = stakeRulesValue
	}

	// x.Costakers (map[string]map[string]uint64->*ast.MapType) is primitive: false
	// template decompose: x.Costakers (map[string]map[string]uint64->*ast.MapType)
	// template map: x.Costakers
	costakersKVs := make([]nt.Value, 0, len(x.Costakers)*2)
	for costakersKey, costakersValue := range x.Costakers {
		// template decompose: costakersValue (map[string]uint64->*ast.MapType)
		// template map: costakersValue
		costakersValueKVs := make([]nt.Value, 0, len(costakersValue)*2)
		for costakersValueKey, costakersValueValue := range costakersValue {
			// template decompose: costakersValueValue (uint64->*ast.Ident)
			costakersValueKVs = append(
				costakersValueKVs,
				nt.String(costakersValueKey),
				util.Int(costakersValueValue).NomsValue(),
			)
		}
		costakersKVs = append(
			costakersKVs,
			nt.String(costakersKey),
			nt.NewMap(vrw, costakersValueKVs...),
		)
	}

	// x.Holds ([]Hold->*ast.ArrayType) is primitive: false
	// template decompose: x.Holds ([]Hold->*ast.ArrayType)
	// template slice: x.Holds
	holdsItems := make([]nt.Value, 0, len(x.Holds))
	for _, holdsItem := range x.Holds {
		// template decompose: holdsItem (Hold->*ast.Ident)
		// template nomsmarshaler: holdsItem
		holdsItemValue, err := holdsItem.MarshalNoms(vrw)
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->holdsItem.MarshalNoms")
		}
		holdsItems = append(
			holdsItems,
			holdsItemValue,
		)
	}

	// x.RecourseSettings (RecourseSettings->*ast.Ident) is primitive: false
	// template decompose: x.RecourseSettings (RecourseSettings->*ast.Ident)
	// template nomsmarshaler: x.RecourseSettings
	recourseSettingsValue, err := x.RecourseSettings.MarshalNoms(vrw)
	if err != nil {
		return nil, errors.Wrap(err, "AccountData.MarshalNoms->RecourseSettings.MarshalNoms")
	}

	// x.CurrencySeatDate (*math.Timestamp->*ast.StarExpr) is primitive: false
	// template decompose: x.CurrencySeatDate (*math.Timestamp->*ast.StarExpr)
	// template pointer:  x.CurrencySeatDate
	var currencySeatDateUnptr nt.Value
	if x.CurrencySeatDate == nil {
		currencySeatDateUnptr = util.Int(0).NomsValue()
	} else {
		// template decompose: (*x.CurrencySeatDate) (math.Timestamp->*ast.SelectorExpr)
		currencySeatDateUnptr = util.Int((*x.CurrencySeatDate)).NomsValue()
	}

	// x.Parent (*address.Address->*ast.StarExpr) is primitive: false
	// template decompose: x.Parent (*address.Address->*ast.StarExpr)
	// template pointer:  x.Parent
	var parentUnptr nt.Value
	if x.Parent == nil {
		parentUnptr = nt.String("")
	} else {
		// template decompose: (*x.Parent) (address.Address->*ast.SelectorExpr)
		// template textmarshaler: (*x.Parent)
		parentString, err := (*x.Parent).MarshalText()
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->Parent.MarshalText")
		}
		parentUnptr = nt.String(parentString)
	}

	// x.Progenitor (*address.Address->*ast.StarExpr) is primitive: false
	// template decompose: x.Progenitor (*address.Address->*ast.StarExpr)
	// template pointer:  x.Progenitor
	var progenitorUnptr nt.Value
	if x.Progenitor == nil {
		progenitorUnptr = nt.String("")
	} else {
		// template decompose: (*x.Progenitor) (address.Address->*ast.SelectorExpr)
		// template textmarshaler: (*x.Progenitor)
		progenitorString, err := (*x.Progenitor).MarshalText()
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->Progenitor.MarshalText")
		}
		progenitorUnptr = nt.String(progenitorString)
	}

	// x.UncreditedEAI (math.Ndau->*ast.SelectorExpr) is primitive: true

	values := []nt.Value{
		// x.Balance (math.Ndau)
		util.Int(x.Balance).NomsValue(),
		// x.Costakers (map[string]map[string]uint64)
		nt.NewMap(vrw, costakersKVs...),
		// x.CurrencySeatDate (*math.Timestamp)
		currencySeatDateUnptr,
		// x.DelegationNode (*address.Address)
		delegationNodeUnptr,
		// x.HasCurrencySeatDate (bool)
		nt.Bool(x.CurrencySeatDate != nil),
		// x.HasDelegationNode (bool)
		nt.Bool(x.DelegationNode != nil),
		// x.HasLock (bool)
		nt.Bool(x.Lock != nil),
		// x.HasParent (bool)
		nt.Bool(x.Parent != nil),
		// x.HasProgenitor (bool)
		nt.Bool(x.Progenitor != nil),
		// x.HasRewardsTarget (bool)
		nt.Bool(x.RewardsTarget != nil),
		// x.HasStakeRules (bool)
		nt.Bool(x.StakeRules != nil),
		// x.Holds ([]Hold)
		nt.NewList(vrw, holdsItems...),
		// x.IncomingRewardsFrom ([]address.Address)
		nt.NewList(vrw, incomingRewardsFromItems...),
		// x.LastEAIUpdate (math.Timestamp)
		util.Int(x.LastEAIUpdate).NomsValue(),
		// x.LastWAAUpdate (math.Timestamp)
		util.Int(x.LastWAAUpdate).NomsValue(),
		// x.Lock (*Lock)
		lockUnptr,
		// x.Parent (*address.Address)
		parentUnptr,
		// x.Progenitor (*address.Address)
		progenitorUnptr,
		// x.RecourseSettings (RecourseSettings)
		recourseSettingsValue,
		// x.RewardsTarget (*address.Address)
		rewardsTargetUnptr,
		// x.Sequence (uint64)
		util.Int(x.Sequence).NomsValue(),
		// x.StakeRules (*StakeRules)
		stakeRulesUnptr,
		// x.UncreditedEAI (math.Ndau)
		util.Int(x.UncreditedEAI).NomsValue(),
		// x.ValidationKeys ([]signature.PublicKey)
		nt.NewList(vrw, validationKeysItems...),
		// x.ValidationScript ([]byte)
		nt.String(x.ValidationScript),
		// x.WeightedAverageAge (math.Duration)
		util.Int(x.WeightedAverageAge).NomsValue(),
	}

	return accountDataStructTemplate.NewStruct(values), nil
}

var _ marshal.Marshaler = (*AccountData)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *AccountData) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"AccountData.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) (stop bool) {
		switch name {
		// x.Balance (math.Ndau->*ast.SelectorExpr) is primitive: true
		case "Balance":
			// template u_decompose: x.Balance (math.Ndau->*ast.SelectorExpr)
			// template u_primitive: x.Balance
			var balanceValue util.Int
			balanceValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "AccountData.UnmarshalNoms->Balance")
				return
			}
			balanceTyped := math.Ndau(balanceValue)

			x.Balance = balanceTyped
		// x.ValidationKeys ([]signature.PublicKey->*ast.ArrayType) is primitive: false
		case "ValidationKeys":
			// template u_decompose: x.ValidationKeys ([]signature.PublicKey->*ast.ArrayType)
			// template u_slice: x.ValidationKeys
			var validationKeysSlice []signature.PublicKey
			if validationKeysList, ok := value.(nt.List); ok {
				validationKeysSlice = make([]signature.PublicKey, 0, validationKeysList.Len())
				validationKeysList.Iter(func(validationKeysItem nt.Value, idx uint64) (stop bool) {

					// template u_decompose: validationKeysItem (signature.PublicKey->*ast.SelectorExpr)
					// template u_textmarshaler: validationKeysItem
					var validationKeysItemValue signature.PublicKey
					if validationKeysItemString, ok := validationKeysItem.(nt.String); ok {
						err = validationKeysItemValue.UnmarshalText([]byte(validationKeysItemString))
					} else {
						err = fmt.Errorf(
							"AccountData.UnmarshalNoms expected validationKeysItem to be a nt.String; found %s",
							reflect.ValueOf(validationKeysItem).Type(),
						)
					}
					if err != nil {
						return true
					}
					validationKeysSlice = append(validationKeysSlice, validationKeysItemValue)
					return false
				})
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms expected value to be a nt.List; found %s",
					reflect.TypeOf(value),
				)
			}

			x.ValidationKeys = validationKeysSlice
		// x.ValidationScript ([]byte->*ast.ArrayType) is primitive: true
		case "ValidationScript":
			// template u_decompose: x.ValidationScript ([]byte->*ast.ArrayType)
			// template u_primitive: x.ValidationScript
			validationScriptValue, ok := value.(nt.String)
			if !ok {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.TypeOf(value),
				)
			}
			validationScriptTyped := []byte(validationScriptValue)

			x.ValidationScript = validationScriptTyped
		// x.RewardsTarget (*address.Address->*ast.StarExpr) is primitive: false
		case "RewardsTarget":
			// template u_decompose: x.RewardsTarget (*address.Address->*ast.StarExpr)
			// template u_pointer:  x.RewardsTarget
			if hasRewardsTargetValue, ok := vs.MaybeGet("HasRewardsTarget"); ok {
				if hasRewardsTarget, ok := hasRewardsTargetValue.(nt.Bool); ok {
					if !hasRewardsTarget {
						return
					}
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms expected HasRewardsTarget to be a nt.Bool; found %s",
						reflect.TypeOf(hasRewardsTargetValue),
					)
					return
				}
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms->RewardsTarget is a pointer, so expected a HasRewardsTarget field: not found",
				)
				return
			}

			// template u_decompose: x.RewardsTarget (address.Address->*ast.SelectorExpr)
			// template u_textmarshaler: x.RewardsTarget
			var rewardsTargetValue address.Address
			if rewardsTargetString, ok := value.(nt.String); ok {
				err = rewardsTargetValue.UnmarshalText([]byte(rewardsTargetString))
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.ValueOf(value).Type(),
				)
			}

			x.RewardsTarget = &rewardsTargetValue
		// x.IncomingRewardsFrom ([]address.Address->*ast.ArrayType) is primitive: false
		case "IncomingRewardsFrom":
			// template u_decompose: x.IncomingRewardsFrom ([]address.Address->*ast.ArrayType)
			// template u_slice: x.IncomingRewardsFrom
			var incomingRewardsFromSlice []address.Address
			if incomingRewardsFromList, ok := value.(nt.List); ok {
				incomingRewardsFromSlice = make([]address.Address, 0, incomingRewardsFromList.Len())
				incomingRewardsFromList.Iter(func(incomingRewardsFromItem nt.Value, idx uint64) (stop bool) {

					// template u_decompose: incomingRewardsFromItem (address.Address->*ast.SelectorExpr)
					// template u_textmarshaler: incomingRewardsFromItem
					var incomingRewardsFromItemValue address.Address
					if incomingRewardsFromItemString, ok := incomingRewardsFromItem.(nt.String); ok {
						err = incomingRewardsFromItemValue.UnmarshalText([]byte(incomingRewardsFromItemString))
					} else {
						err = fmt.Errorf(
							"AccountData.UnmarshalNoms expected incomingRewardsFromItem to be a nt.String; found %s",
							reflect.ValueOf(incomingRewardsFromItem).Type(),
						)
					}
					if err != nil {
						return true
					}
					incomingRewardsFromSlice = append(incomingRewardsFromSlice, incomingRewardsFromItemValue)
					return false
				})
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms expected value to be a nt.List; found %s",
					reflect.TypeOf(value),
				)
			}

			x.IncomingRewardsFrom = incomingRewardsFromSlice
		// x.DelegationNode (*address.Address->*ast.StarExpr) is primitive: false
		case "DelegationNode":
			// template u_decompose: x.DelegationNode (*address.Address->*ast.StarExpr)
			// template u_pointer:  x.DelegationNode
			if hasDelegationNodeValue, ok := vs.MaybeGet("HasDelegationNode"); ok {
				if hasDelegationNode, ok := hasDelegationNodeValue.(nt.Bool); ok {
					if !hasDelegationNode {
						return
					}
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms expected HasDelegationNode to be a nt.Bool; found %s",
						reflect.TypeOf(hasDelegationNodeValue),
					)
					return
				}
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms->DelegationNode is a pointer, so expected a HasDelegationNode field: not found",
				)
				return
			}

			// template u_decompose: x.DelegationNode (address.Address->*ast.SelectorExpr)
			// template u_textmarshaler: x.DelegationNode
			var delegationNodeValue address.Address
			if delegationNodeString, ok := value.(nt.String); ok {
				err = delegationNodeValue.UnmarshalText([]byte(delegationNodeString))
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.ValueOf(value).Type(),
				)
			}

			x.DelegationNode = &delegationNodeValue
		// x.Lock (*Lock->*ast.StarExpr) is primitive: false
		case "Lock":
			// template u_decompose: x.Lock (*Lock->*ast.StarExpr)
			// template u_pointer:  x.Lock
			if hasLockValue, ok := vs.MaybeGet("HasLock"); ok {
				if hasLock, ok := hasLockValue.(nt.Bool); ok {
					if !hasLock {
						return
					}
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms expected HasLock to be a nt.Bool; found %s",
						reflect.TypeOf(hasLockValue),
					)
					return
				}
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms->Lock is a pointer, so expected a HasLock field: not found",
				)
				return
			}

			// template u_decompose: x.Lock (Lock->*ast.Ident)
			// template u_nomsmarshaler: x.Lock
			var lockInstance Lock
			err = lockInstance.UnmarshalNoms(value)
			err = errors.Wrap(err, "AccountData.UnmarshalNoms->Lock")

			x.Lock = &lockInstance
		// x.LastEAIUpdate (math.Timestamp->*ast.SelectorExpr) is primitive: true
		case "LastEAIUpdate":
			// template u_decompose: x.LastEAIUpdate (math.Timestamp->*ast.SelectorExpr)
			// template u_primitive: x.LastEAIUpdate
			var lastEAIUpdateValue util.Int
			lastEAIUpdateValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "AccountData.UnmarshalNoms->LastEAIUpdate")
				return
			}
			lastEAIUpdateTyped := math.Timestamp(lastEAIUpdateValue)

			x.LastEAIUpdate = lastEAIUpdateTyped
		// x.LastWAAUpdate (math.Timestamp->*ast.SelectorExpr) is primitive: true
		case "LastWAAUpdate":
			// template u_decompose: x.LastWAAUpdate (math.Timestamp->*ast.SelectorExpr)
			// template u_primitive: x.LastWAAUpdate
			var lastWAAUpdateValue util.Int
			lastWAAUpdateValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "AccountData.UnmarshalNoms->LastWAAUpdate")
				return
			}
			lastWAAUpdateTyped := math.Timestamp(lastWAAUpdateValue)

			x.LastWAAUpdate = lastWAAUpdateTyped
		// x.WeightedAverageAge (math.Duration->*ast.SelectorExpr) is primitive: true
		case "WeightedAverageAge":
			// template u_decompose: x.WeightedAverageAge (math.Duration->*ast.SelectorExpr)
			// template u_primitive: x.WeightedAverageAge
			var weightedAverageAgeValue util.Int
			weightedAverageAgeValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "AccountData.UnmarshalNoms->WeightedAverageAge")
				return
			}
			weightedAverageAgeTyped := math.Duration(weightedAverageAgeValue)

			x.WeightedAverageAge = weightedAverageAgeTyped
		// x.Sequence (uint64->*ast.Ident) is primitive: true
		case "Sequence":
			// template u_decompose: x.Sequence (uint64->*ast.Ident)
			// template u_primitive: x.Sequence
			var sequenceValue util.Int
			sequenceValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "AccountData.UnmarshalNoms->Sequence")
				return
			}
			sequenceTyped := uint64(sequenceValue)

			x.Sequence = sequenceTyped
		// x.StakeRules (*StakeRules->*ast.StarExpr) is primitive: false
		case "StakeRules":
			// template u_decompose: x.StakeRules (*StakeRules->*ast.StarExpr)
			// template u_pointer:  x.StakeRules
			if hasStakeRulesValue, ok := vs.MaybeGet("HasStakeRules"); ok {
				if hasStakeRules, ok := hasStakeRulesValue.(nt.Bool); ok {
					if !hasStakeRules {
						return
					}
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms expected HasStakeRules to be a nt.Bool; found %s",
						reflect.TypeOf(hasStakeRulesValue),
					)
					return
				}
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms->StakeRules is a pointer, so expected a HasStakeRules field: not found",
				)
				return
			}

			// template u_decompose: x.StakeRules (StakeRules->*ast.Ident)
			// template u_nomsmarshaler: x.StakeRules
			var stakeRulesInstance StakeRules
			err = stakeRulesInstance.UnmarshalNoms(value)
			err = errors.Wrap(err, "AccountData.UnmarshalNoms->StakeRules")

			x.StakeRules = &stakeRulesInstance
		// x.Costakers (map[string]map[string]uint64->*ast.MapType) is primitive: false
		case "Costakers":
			// template u_decompose: x.Costakers (map[string]map[string]uint64->*ast.MapType)
			// template u_map: x.Costakers
			costakersGMap := make(map[string]map[string]uint64)
			if costakersNMap, ok := value.(nt.Map); ok {
				costakersNMap.Iter(func(costakersKey, costakersValue nt.Value) (stop bool) {
					costakersKeyString, ok := costakersKey.(nt.String)
					if !ok {
						err = fmt.Errorf(
							"AccountData.UnmarshalNoms expected costakersKey to be a nt.String; found %s",
							reflect.TypeOf(costakersKey),
						)
						return true
					}

					// template u_decompose: costakersValue (map[string]uint64->*ast.MapType)
					// template u_map: costakersValue
					costakersValueGMap := make(map[string]uint64)
					if costakersValueNMap, ok := costakersValue.(nt.Map); ok {
						costakersValueNMap.Iter(func(costakersValueKey, costakersValueValue nt.Value) (stop bool) {
							costakersValueKeyString, ok := costakersValueKey.(nt.String)
							if !ok {
								err = fmt.Errorf(
									"AccountData.UnmarshalNoms expected costakersValueKey to be a nt.String; found %s",
									reflect.TypeOf(costakersValueKey),
								)
								return true
							}

							// template u_decompose: costakersValueValue (uint64->*ast.Ident)
							// template u_primitive: costakersValueValue
							var costakersValueValueValue util.Int
							costakersValueValueValue, err = util.IntFrom(costakersValueValue)
							if err != nil {
								err = errors.Wrap(err, "AccountData.UnmarshalNoms->costakersValueValue")
								return
							}
							costakersValueValueTyped := uint64(costakersValueValueValue)
							if err != nil {
								return true
							}
							costakersValueGMap[string(costakersValueKeyString)] = costakersValueValueTyped
							return false
						})
					} else {
						err = fmt.Errorf(
							"AccountData.UnmarshalNoms expected costakersValueGMap to be a nt.Map; found %s",
							reflect.TypeOf(costakersValue),
						)
					}
					if err != nil {
						return true
					}
					costakersGMap[string(costakersKeyString)] = costakersValueGMap
					return false
				})
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms expected costakersGMap to be a nt.Map; found %s",
					reflect.TypeOf(value),
				)
			}

			x.Costakers = costakersGMap
		// x.Holds ([]Hold->*ast.ArrayType) is primitive: false
		case "Holds":
			// template u_decompose: x.Holds ([]Hold->*ast.ArrayType)
			// template u_slice: x.Holds
			var holdsSlice []Hold
			if holdsList, ok := value.(nt.List); ok {
				holdsSlice = make([]Hold, 0, holdsList.Len())
				holdsList.Iter(func(holdsItem nt.Value, idx uint64) (stop bool) {

					// template u_decompose: holdsItem (Hold->*ast.Ident)
					// template u_nomsmarshaler: holdsItem
					var holdsItemInstance Hold
					err = holdsItemInstance.UnmarshalNoms(holdsItem)
					err = errors.Wrap(err, "AccountData.UnmarshalNoms->holdsItem")
					if err != nil {
						return true
					}
					holdsSlice = append(holdsSlice, holdsItemInstance)
					return false
				})
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms expected value to be a nt.List; found %s",
					reflect.TypeOf(value),
				)
			}

			x.Holds = holdsSlice
		// x.RecourseSettings (RecourseSettings->*ast.Ident) is primitive: false
		case "RecourseSettings":
			// template u_decompose: x.RecourseSettings (RecourseSettings->*ast.Ident)
			// template u_nomsmarshaler: x.RecourseSettings
			var recourseSettingsInstance RecourseSettings
			err = recourseSettingsInstance.UnmarshalNoms(value)
			err = errors.Wrap(err, "AccountData.UnmarshalNoms->RecourseSettings")

			x.RecourseSettings = recourseSettingsInstance
		// x.CurrencySeatDate (*math.Timestamp->*ast.StarExpr) is primitive: false
		case "CurrencySeatDate":
			// template u_decompose: x.CurrencySeatDate (*math.Timestamp->*ast.StarExpr)
			// template u_pointer:  x.CurrencySeatDate
			if hasCurrencySeatDateValue, ok := vs.MaybeGet("HasCurrencySeatDate"); ok {
				if hasCurrencySeatDate, ok := hasCurrencySeatDateValue.(nt.Bool); ok {
					if !hasCurrencySeatDate {
						return
					}
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms expected HasCurrencySeatDate to be a nt.Bool; found %s",
						reflect.TypeOf(hasCurrencySeatDateValue),
					)
					return
				}
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms->CurrencySeatDate is a pointer, so expected a HasCurrencySeatDate field: not found",
				)
				return
			}

			// template u_decompose: x.CurrencySeatDate (math.Timestamp->*ast.SelectorExpr)
			// template u_primitive: x.CurrencySeatDate
			var currencySeatDateValue util.Int
			currencySeatDateValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "AccountData.UnmarshalNoms->CurrencySeatDate")
				return
			}
			currencySeatDateTyped := math.Timestamp(currencySeatDateValue)

			x.CurrencySeatDate = &currencySeatDateTyped
		// x.Parent (*address.Address->*ast.StarExpr) is primitive: false
		case "Parent":
			// template u_decompose: x.Parent (*address.Address->*ast.StarExpr)
			// template u_pointer:  x.Parent
			if hasParentValue, ok := vs.MaybeGet("HasParent"); ok {
				if hasParent, ok := hasParentValue.(nt.Bool); ok {
					if !hasParent {
						return
					}
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms expected HasParent to be a nt.Bool; found %s",
						reflect.TypeOf(hasParentValue),
					)
					return
				}
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms->Parent is a pointer, so expected a HasParent field: not found",
				)
				return
			}

			// template u_decompose: x.Parent (address.Address->*ast.SelectorExpr)
			// template u_textmarshaler: x.Parent
			var parentValue address.Address
			if parentString, ok := value.(nt.String); ok {
				err = parentValue.UnmarshalText([]byte(parentString))
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.ValueOf(value).Type(),
				)
			}

			x.Parent = &parentValue
		// x.Progenitor (*address.Address->*ast.StarExpr) is primitive: false
		case "Progenitor":
			// template u_decompose: x.Progenitor (*address.Address->*ast.StarExpr)
			// template u_pointer:  x.Progenitor
			if hasProgenitorValue, ok := vs.MaybeGet("HasProgenitor"); ok {
				if hasProgenitor, ok := hasProgenitorValue.(nt.Bool); ok {
					if !hasProgenitor {
						return
					}
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms expected HasProgenitor to be a nt.Bool; found %s",
						reflect.TypeOf(hasProgenitorValue),
					)
					return
				}
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms->Progenitor is a pointer, so expected a HasProgenitor field: not found",
				)
				return
			}

			// template u_decompose: x.Progenitor (address.Address->*ast.SelectorExpr)
			// template u_textmarshaler: x.Progenitor
			var progenitorValue address.Address
			if progenitorString, ok := value.(nt.String); ok {
				err = progenitorValue.UnmarshalText([]byte(progenitorString))
			} else {
				err = fmt.Errorf(
					"AccountData.UnmarshalNoms expected value to be a nt.String; found %s",
					reflect.ValueOf(value).Type(),
				)
			}

			x.Progenitor = &progenitorValue
		// x.UncreditedEAI (math.Ndau->*ast.SelectorExpr) is primitive: true
		case "UncreditedEAI":
			// template u_decompose: x.UncreditedEAI (math.Ndau->*ast.SelectorExpr)
			// template u_primitive: x.UncreditedEAI
			var uncreditedEAIValue util.Int
			uncreditedEAIValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "AccountData.UnmarshalNoms->UncreditedEAI")
				return
			}
			uncreditedEAITyped := math.Ndau(uncreditedEAIValue)

			x.UncreditedEAI = uncreditedEAITyped
		}
		stop = err != nil
		return
	})
	return
}

var _ marshal.Unmarshaler = (*AccountData)(nil)

var recourseSettingsStructTemplate nt.StructTemplate

func init() {
	recourseSettingsStructTemplate = nt.MakeStructTemplate("RecourseSettings", []string{
		"ChangesAt",
		"HasChangesAt",
		"HasNext",
		"Next",
		"Period",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x RecourseSettings) MarshalNoms(vrw nt.ValueReadWriter) (recourseSettingsValue nt.Value, err error) {
	// x.Period (math.Duration->*ast.SelectorExpr) is primitive: true

	// x.ChangesAt (*math.Timestamp->*ast.StarExpr) is primitive: false
	// template decompose: x.ChangesAt (*math.Timestamp->*ast.StarExpr)
	// template pointer:  x.ChangesAt
	var changesAtUnptr nt.Value
	if x.ChangesAt == nil {
		changesAtUnptr = util.Int(0).NomsValue()
	} else {
		// template decompose: (*x.ChangesAt) (math.Timestamp->*ast.SelectorExpr)
		changesAtUnptr = util.Int((*x.ChangesAt)).NomsValue()
	}

	// x.Next (*math.Duration->*ast.StarExpr) is primitive: false
	// template decompose: x.Next (*math.Duration->*ast.StarExpr)
	// template pointer:  x.Next
	var nextUnptr nt.Value
	if x.Next == nil {
		nextUnptr = util.Int(0).NomsValue()
	} else {
		// template decompose: (*x.Next) (math.Duration->*ast.SelectorExpr)
		nextUnptr = util.Int((*x.Next)).NomsValue()
	}

	values := []nt.Value{
		// x.ChangesAt (*math.Timestamp)
		changesAtUnptr,
		// x.HasChangesAt (bool)
		nt.Bool(x.ChangesAt != nil),
		// x.HasNext (bool)
		nt.Bool(x.Next != nil),
		// x.Next (*math.Duration)
		nextUnptr,
		// x.Period (math.Duration)
		util.Int(x.Period).NomsValue(),
	}

	return recourseSettingsStructTemplate.NewStruct(values), nil
}

var _ marshal.Marshaler = (*RecourseSettings)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *RecourseSettings) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"RecourseSettings.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) (stop bool) {
		switch name {
		// x.Period (math.Duration->*ast.SelectorExpr) is primitive: true
		case "Period":
			// template u_decompose: x.Period (math.Duration->*ast.SelectorExpr)
			// template u_primitive: x.Period
			var periodValue util.Int
			periodValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "RecourseSettings.UnmarshalNoms->Period")
				return
			}
			periodTyped := math.Duration(periodValue)

			x.Period = periodTyped
		// x.ChangesAt (*math.Timestamp->*ast.StarExpr) is primitive: false
		case "ChangesAt":
			// template u_decompose: x.ChangesAt (*math.Timestamp->*ast.StarExpr)
			// template u_pointer:  x.ChangesAt
			if hasChangesAtValue, ok := vs.MaybeGet("HasChangesAt"); ok {
				if hasChangesAt, ok := hasChangesAtValue.(nt.Bool); ok {
					if !hasChangesAt {
						return
					}
				} else {
					err = fmt.Errorf(
						"RecourseSettings.UnmarshalNoms expected HasChangesAt to be a nt.Bool; found %s",
						reflect.TypeOf(hasChangesAtValue),
					)
					return
				}
			} else {
				err = fmt.Errorf(
					"RecourseSettings.UnmarshalNoms->ChangesAt is a pointer, so expected a HasChangesAt field: not found",
				)
				return
			}

			// template u_decompose: x.ChangesAt (math.Timestamp->*ast.SelectorExpr)
			// template u_primitive: x.ChangesAt
			var changesAtValue util.Int
			changesAtValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "RecourseSettings.UnmarshalNoms->ChangesAt")
				return
			}
			changesAtTyped := math.Timestamp(changesAtValue)

			x.ChangesAt = &changesAtTyped
		// x.Next (*math.Duration->*ast.StarExpr) is primitive: false
		case "Next":
			// template u_decompose: x.Next (*math.Duration->*ast.StarExpr)
			// template u_pointer:  x.Next
			if hasNextValue, ok := vs.MaybeGet("HasNext"); ok {
				if hasNext, ok := hasNextValue.(nt.Bool); ok {
					if !hasNext {
						return
					}
				} else {
					err = fmt.Errorf(
						"RecourseSettings.UnmarshalNoms expected HasNext to be a nt.Bool; found %s",
						reflect.TypeOf(hasNextValue),
					)
					return
				}
			} else {
				err = fmt.Errorf(
					"RecourseSettings.UnmarshalNoms->Next is a pointer, so expected a HasNext field: not found",
				)
				return
			}

			// template u_decompose: x.Next (math.Duration->*ast.SelectorExpr)
			// template u_primitive: x.Next
			var nextValue util.Int
			nextValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "RecourseSettings.UnmarshalNoms->Next")
				return
			}
			nextTyped := math.Duration(nextValue)

			x.Next = &nextTyped
		}
		stop = err != nil
		return
	})
	return
}

var _ marshal.Unmarshaler = (*RecourseSettings)(nil)
