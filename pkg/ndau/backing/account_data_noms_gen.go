package backing

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
)

// this code generated by github.com/oneiro-ndev/generator/cmd/nomsify
// DO NOT EDIT

var accountDataStructTemplate nt.StructTemplate

func init() {
	accountDataStructTemplate = nt.MakeStructTemplate("AccountData", []string{
		"Balance",
		"CurrencySeatDate",
		"DelegationNode",
		"HasCurrencySeatDate",
		"HasDelegationNode",
		"HasLock",
		"HasParent",
		"HasProgenitor",
		"HasRewardsTarget",
		"HasStake",
		"IncomingRewardsFrom",
		"LastEAIUpdate",
		"LastWAAUpdate",
		"Lock",
		"Parent",
		"Progenitor",
		"RewardsTarget",
		"Sequence",
		"SettlementSettings",
		"Settlements",
		"SidechainPayments",
		"Stake",
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

	// x.Stake (*Stake->*ast.StarExpr) is primitive: false

	// template decompose: x.Stake (*Stake->*ast.StarExpr)
	// template pointer:  x.Stake
	var stakeUnptr nt.Value
	if x.Stake == nil {
		stakeUnptr = nt.Bool(false)
	} else {

		// template decompose: (*x.Stake) (Stake->*ast.Ident)
		// template nomsmarshaler: (*x.Stake)
		stakeValue, err := (*x.Stake).MarshalNoms(vrw)
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->Stake.MarshalNoms")
		}

		stakeUnptr = stakeValue
	}

	// x.LastEAIUpdate (math.Timestamp->*ast.SelectorExpr) is primitive: true

	// x.LastWAAUpdate (math.Timestamp->*ast.SelectorExpr) is primitive: true

	// x.WeightedAverageAge (math.Duration->*ast.SelectorExpr) is primitive: true

	// x.Sequence (uint64->*ast.Ident) is primitive: true

	// x.Settlements ([]Settlement->*ast.ArrayType) is primitive: false

	// template decompose: x.Settlements ([]Settlement->*ast.ArrayType)
	// template slice: x.Settlements
	settlementsItems := make([]nt.Value, 0, len(x.Settlements))
	for _, settlementsItem := range x.Settlements {

		// template decompose: settlementsItem (Settlement->*ast.Ident)
		// template nomsmarshaler: settlementsItem
		settlementsItemValue, err := settlementsItem.MarshalNoms(vrw)
		if err != nil {
			return nil, errors.Wrap(err, "AccountData.MarshalNoms->settlementsItem.MarshalNoms")
		}

		settlementsItems = append(
			settlementsItems,
			settlementsItemValue,
		)
	}

	// x.SettlementSettings (SettlementSettings->*ast.Ident) is primitive: false

	// template decompose: x.SettlementSettings (SettlementSettings->*ast.Ident)
	// template nomsmarshaler: x.SettlementSettings
	settlementSettingsValue, err := x.SettlementSettings.MarshalNoms(vrw)
	if err != nil {
		return nil, errors.Wrap(err, "AccountData.MarshalNoms->SettlementSettings.MarshalNoms")
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

	// x.SidechainPayments (map[string]struct{}->*ast.MapType) is primitive: false

	// template decompose: x.SidechainPayments (map[string]struct{}->*ast.MapType)
	// template set:  x.SidechainPayments
	sidechainPaymentsItems := make([]nt.Value, 0, len(x.SidechainPayments))
	if len(x.SidechainPayments) > 0 {
		// We need to iterate the set in sorted order, so build []string and sort it first
		sidechainPaymentsSorted := make([]string, 0, len(x.SidechainPayments))
		for sidechainPaymentsItem := range x.SidechainPayments {
			sidechainPaymentsSorted = append(sidechainPaymentsSorted, sidechainPaymentsItem)
		}
		sort.Sort(sort.StringSlice(sidechainPaymentsSorted))
		for _, sidechainPaymentsItem := range sidechainPaymentsSorted {
			sidechainPaymentsItems = append(
				sidechainPaymentsItems,
				nt.String(sidechainPaymentsItem),
			)
		}
	}

	// x.UncreditedEAI (math.Ndau->*ast.SelectorExpr) is primitive: true

	return accountDataStructTemplate.NewStruct([]nt.Value{
		// x.Balance (math.Ndau)

		util.Int(x.Balance).NomsValue(),
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
		// x.HasStake (bool)

		nt.Bool(x.Stake != nil),
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
		// x.RewardsTarget (*address.Address)
		rewardsTargetUnptr,
		// x.Sequence (uint64)

		util.Int(x.Sequence).NomsValue(),
		// x.SettlementSettings (SettlementSettings)
		settlementSettingsValue,
		// x.Settlements ([]Settlement)
		nt.NewList(vrw, settlementsItems...),
		// x.SidechainPayments (map[string]struct{})
		nt.NewSet(vrw, sidechainPaymentsItems...),
		// x.Stake (*Stake)
		stakeUnptr,
		// x.UncreditedEAI (math.Ndau)

		util.Int(x.UncreditedEAI).NomsValue(),
		// x.ValidationKeys ([]signature.PublicKey)
		nt.NewList(vrw, validationKeysItems...),
		// x.ValidationScript ([]byte)

		nt.String(x.ValidationScript),
		// x.WeightedAverageAge (math.Duration)

		util.Int(x.WeightedAverageAge).NomsValue(),
	}), nil
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
	vs.IterFields(func(name string, value nt.Value) {
		if err == nil {
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
				if validationKeysList, ok := value.(nt.List); ok {
					x.ValidationKeys = make([]signature.PublicKey, 0, validationKeysList.Len())
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
						x.ValidationKeys = append(x.ValidationKeys, validationKeysItemValue)
						return false
					})
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms expected x.ValidationKeys to be a nt.List; found %s",
						reflect.TypeOf(value),
					)
				}

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
				if incomingRewardsFromList, ok := value.(nt.List); ok {
					x.IncomingRewardsFrom = make([]address.Address, 0, incomingRewardsFromList.Len())
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
						x.IncomingRewardsFrom = append(x.IncomingRewardsFrom, incomingRewardsFromItemValue)
						return false
					})
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms expected x.IncomingRewardsFrom to be a nt.List; found %s",
						reflect.TypeOf(value),
					)
				}

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

			// x.Stake (*Stake->*ast.StarExpr) is primitive: false
			case "Stake":
				// template u_decompose: x.Stake (*Stake->*ast.StarExpr)
				// template u_pointer:  x.Stake
				if hasStakeValue, ok := vs.MaybeGet("HasStake"); ok {
					if hasStake, ok := hasStakeValue.(nt.Bool); ok {
						if !hasStake {
							return
						}
					} else {
						err = fmt.Errorf(
							"AccountData.UnmarshalNoms expected HasStake to be a nt.Bool; found %s",
							reflect.TypeOf(hasStakeValue),
						)
						return
					}
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms->Stake is a pointer, so expected a HasStake field: not found",
					)
					return
				}

				// template u_decompose: x.Stake (Stake->*ast.Ident)
				// template u_nomsmarshaler: x.Stake
				var stakeInstance Stake
				err = stakeInstance.UnmarshalNoms(value)
				err = errors.Wrap(err, "AccountData.UnmarshalNoms->Stake")

				x.Stake = &stakeInstance

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

			// x.Settlements ([]Settlement->*ast.ArrayType) is primitive: false
			case "Settlements":
				// template u_decompose: x.Settlements ([]Settlement->*ast.ArrayType)
				// template u_slice: x.Settlements
				if settlementsList, ok := value.(nt.List); ok {
					x.Settlements = make([]Settlement, 0, settlementsList.Len())
					settlementsList.Iter(func(settlementsItem nt.Value, idx uint64) (stop bool) {

						// template u_decompose: settlementsItem (Settlement->*ast.Ident)
						// template u_nomsmarshaler: settlementsItem
						var settlementsItemInstance Settlement
						err = settlementsItemInstance.UnmarshalNoms(settlementsItem)
						err = errors.Wrap(err, "AccountData.UnmarshalNoms->settlementsItem")
						if err != nil {
							return true
						}
						x.Settlements = append(x.Settlements, settlementsItemInstance)
						return false
					})
				} else {
					err = fmt.Errorf(
						"AccountData.UnmarshalNoms expected x.Settlements to be a nt.List; found %s",
						reflect.TypeOf(value),
					)
				}

			// x.SettlementSettings (SettlementSettings->*ast.Ident) is primitive: false
			case "SettlementSettings":
				// template u_decompose: x.SettlementSettings (SettlementSettings->*ast.Ident)
				// template u_nomsmarshaler: x.SettlementSettings
				var settlementSettingsInstance SettlementSettings
				err = settlementSettingsInstance.UnmarshalNoms(value)
				err = errors.Wrap(err, "AccountData.UnmarshalNoms->SettlementSettings")

				x.SettlementSettings = settlementSettingsInstance

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

			// x.SidechainPayments (map[string]struct{}->*ast.MapType) is primitive: false
			case "SidechainPayments":
				// template u_decompose: x.SidechainPayments (map[string]struct{}->*ast.MapType)
				// template u_set: x.SidechainPayments
				x.SidechainPayments = make(map[string]struct{})
				if sidechainPaymentsSet, ok := value.(nt.Set); ok {
					sidechainPaymentsSet.Iter(func(sidechainPaymentsItem nt.Value) (stop bool) {
						if sidechainPaymentsItemString, ok := sidechainPaymentsItem.(nt.String); ok {
							x.SidechainPayments[string(sidechainPaymentsItemString)] = struct{}{}
						} else {
							err = fmt.Errorf(
								"AccountData.AccountData.UnmarshalNoms expected SidechainPaymentsItem to be a nt.String; found %s",
								reflect.TypeOf(value),
							)
						}
						return err != nil
					})
				} else {
					err = fmt.Errorf(
						"AccountData.AccountData.UnmarshalNoms expected SidechainPayments to be a nt.Set; found %s",
						reflect.TypeOf(value),
					)
				}

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
		}
	})
	return
}

var _ marshal.Unmarshaler = (*AccountData)(nil)

var stakeStructTemplate nt.StructTemplate

func init() {
	stakeStructTemplate = nt.MakeStructTemplate("Stake", []string{
		"Address",
		"Point",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x Stake) MarshalNoms(vrw nt.ValueReadWriter) (stakeValue nt.Value, err error) {
	// x.Point (math.Timestamp->*ast.SelectorExpr) is primitive: true

	// x.Address (address.Address->*ast.SelectorExpr) is primitive: false

	// template decompose: x.Address (address.Address->*ast.SelectorExpr)
	// template textmarshaler: x.Address
	addressString, err := x.Address.MarshalText()
	if err != nil {
		return nil, errors.Wrap(err, "Stake.MarshalNoms->Address.MarshalText")
	}

	return stakeStructTemplate.NewStruct([]nt.Value{
		// x.Address (address.Address)
		nt.String(addressString),
		// x.Point (math.Timestamp)

		util.Int(x.Point).NomsValue(),
	}), nil
}

var _ marshal.Marshaler = (*Stake)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *Stake) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"Stake.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) {
		if err == nil {
			switch name {
			// x.Point (math.Timestamp->*ast.SelectorExpr) is primitive: true
			case "Point":
				// template u_decompose: x.Point (math.Timestamp->*ast.SelectorExpr)
				// template u_primitive: x.Point
				var pointValue util.Int
				pointValue, err = util.IntFrom(value)
				if err != nil {
					err = errors.Wrap(err, "Stake.UnmarshalNoms->Point")
					return
				}
				pointTyped := math.Timestamp(pointValue)

				x.Point = pointTyped

			// x.Address (address.Address->*ast.SelectorExpr) is primitive: false
			case "Address":
				// template u_decompose: x.Address (address.Address->*ast.SelectorExpr)
				// template u_textmarshaler: x.Address
				var addressValue address.Address
				if addressString, ok := value.(nt.String); ok {
					err = addressValue.UnmarshalText([]byte(addressString))
				} else {
					err = fmt.Errorf(
						"Stake.UnmarshalNoms expected value to be a nt.String; found %s",
						reflect.ValueOf(value).Type(),
					)
				}

				x.Address = addressValue

			}
		}
	})
	return
}

var _ marshal.Unmarshaler = (*Stake)(nil)

var settlementStructTemplate nt.StructTemplate

func init() {
	settlementStructTemplate = nt.MakeStructTemplate("Settlement", []string{
		"Expiry",
		"Qty",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x Settlement) MarshalNoms(vrw nt.ValueReadWriter) (settlementValue nt.Value, err error) {
	// x.Qty (math.Ndau->*ast.SelectorExpr) is primitive: true

	// x.Expiry (math.Timestamp->*ast.SelectorExpr) is primitive: true

	return settlementStructTemplate.NewStruct([]nt.Value{
		// x.Expiry (math.Timestamp)

		util.Int(x.Expiry).NomsValue(),
		// x.Qty (math.Ndau)

		util.Int(x.Qty).NomsValue(),
	}), nil
}

var _ marshal.Marshaler = (*Settlement)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *Settlement) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"Settlement.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) {
		if err == nil {
			switch name {
			// x.Qty (math.Ndau->*ast.SelectorExpr) is primitive: true
			case "Qty":
				// template u_decompose: x.Qty (math.Ndau->*ast.SelectorExpr)
				// template u_primitive: x.Qty
				var qtyValue util.Int
				qtyValue, err = util.IntFrom(value)
				if err != nil {
					err = errors.Wrap(err, "Settlement.UnmarshalNoms->Qty")
					return
				}
				qtyTyped := math.Ndau(qtyValue)

				x.Qty = qtyTyped

			// x.Expiry (math.Timestamp->*ast.SelectorExpr) is primitive: true
			case "Expiry":
				// template u_decompose: x.Expiry (math.Timestamp->*ast.SelectorExpr)
				// template u_primitive: x.Expiry
				var expiryValue util.Int
				expiryValue, err = util.IntFrom(value)
				if err != nil {
					err = errors.Wrap(err, "Settlement.UnmarshalNoms->Expiry")
					return
				}
				expiryTyped := math.Timestamp(expiryValue)

				x.Expiry = expiryTyped

			}
		}
	})
	return
}

var _ marshal.Unmarshaler = (*Settlement)(nil)

var settlementSettingsStructTemplate nt.StructTemplate

func init() {
	settlementSettingsStructTemplate = nt.MakeStructTemplate("SettlementSettings", []string{
		"ChangesAt",
		"HasChangesAt",
		"HasNext",
		"Next",
		"Period",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x SettlementSettings) MarshalNoms(vrw nt.ValueReadWriter) (settlementSettingsValue nt.Value, err error) {
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

	return settlementSettingsStructTemplate.NewStruct([]nt.Value{
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
	}), nil
}

var _ marshal.Marshaler = (*SettlementSettings)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *SettlementSettings) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"SettlementSettings.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) {
		if err == nil {
			switch name {
			// x.Period (math.Duration->*ast.SelectorExpr) is primitive: true
			case "Period":
				// template u_decompose: x.Period (math.Duration->*ast.SelectorExpr)
				// template u_primitive: x.Period
				var periodValue util.Int
				periodValue, err = util.IntFrom(value)
				if err != nil {
					err = errors.Wrap(err, "SettlementSettings.UnmarshalNoms->Period")
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
							"SettlementSettings.UnmarshalNoms expected HasChangesAt to be a nt.Bool; found %s",
							reflect.TypeOf(hasChangesAtValue),
						)
						return
					}
				} else {
					err = fmt.Errorf(
						"SettlementSettings.UnmarshalNoms->ChangesAt is a pointer, so expected a HasChangesAt field: not found",
					)
					return
				}

				// template u_decompose: x.ChangesAt (math.Timestamp->*ast.SelectorExpr)
				// template u_primitive: x.ChangesAt
				var changesAtValue util.Int
				changesAtValue, err = util.IntFrom(value)
				if err != nil {
					err = errors.Wrap(err, "SettlementSettings.UnmarshalNoms->ChangesAt")
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
							"SettlementSettings.UnmarshalNoms expected HasNext to be a nt.Bool; found %s",
							reflect.TypeOf(hasNextValue),
						)
						return
					}
				} else {
					err = fmt.Errorf(
						"SettlementSettings.UnmarshalNoms->Next is a pointer, so expected a HasNext field: not found",
					)
					return
				}

				// template u_decompose: x.Next (math.Duration->*ast.SelectorExpr)
				// template u_primitive: x.Next
				var nextValue util.Int
				nextValue, err = util.IntFrom(value)
				if err != nil {
					err = errors.Wrap(err, "SettlementSettings.UnmarshalNoms->Next")
					return
				}
				nextTyped := math.Duration(nextValue)

				x.Next = &nextTyped

			}
		}
	})
	return
}

var _ marshal.Unmarshaler = (*SettlementSettings)(nil)
