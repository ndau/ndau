package ndau

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"sort"

	"github.com/ndau/msgp-well-known-types/wkt"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	math "github.com/ndau/ndaumath/pkg/types"
	"github.com/ndau/ndaumath/pkg/unsigned"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (app *App) goodnessOf(addrS string) (int64, error) {
	state := app.GetState().(*backing.State)

	addr, err := address.Validate(addrS)
	if err != nil {
		return 0, err
	}

	acct, hasAcct := app.getAccount(addr)
	if !hasAcct {
		return 0, errors.New("no such account")
	}

	node, hasNode := state.Nodes[addrS]
	if !hasNode {
		return 0, errors.New("no such node")
	}

	var script wkt.Bytes
	err = app.System(sv.NodeGoodnessFuncName, &script)
	if err != nil {
		return 0, errors.Wrap(err, "getting goodness script")
	}

	totalStake := math.Ndau(0)
	costakers, err := app.NodeStakers(addr)
	if err != nil {
		return 0, errors.Wrap(err, "getting node stakers")
	}
	for _, v := range costakers {
		totalStake += v
	}

	vm, err := BuildVMForNodeGoodness(
		script,
		addr,
		node.TMAddress,
		acct,
		totalStake,
		app.BlockTime(),
		app.GetStats(),
		node,
		app,
	)
	if err != nil {
		return 0, errors.Wrap(err, "building goodness vm")
	}

	err = vm.Run(nil)
	if err != nil {
		return 0, errors.Wrap(err, "running goodness vm")
	}

	goodness, err := vm.Stack().PopAsInt64()
	if err != nil {
		return goodness, errors.Wrap(err, "goodness stack top not numeric")
	}

	// dump goodness value after chaincode is called
	logger := app.DecoratedLogger().WithFields(log.Fields{
		"address": addr,
		"value":   goodness,
	})
	logger.Info("nodegoodness value")

	return goodness, nil
}

type goodnessPair struct {
	addr     string
	goodness uint64
}

func nodeGoodnesses(app *App) ([]goodnessPair, uint64) {
	state := app.GetState().(*backing.State)
	var goodnessSum uint64
	goodnesses := make([]goodnessPair, 0, len(state.Nodes))
	for addr, node := range state.Nodes {
		if !node.Active {
			continue
		}
		goodness, err := app.goodnessFunc(addr)
		// Remove the test that considers a goodness of 0 to be an error, so
		// nodes can be dropped from the validator set by setting their goodness to 0.
		if err == nil {
			goodnessSum += uint64(goodness)
			goodnesses = append(
				goodnesses,
				goodnessPair{
					addr:     addr,
					goodness: uint64(goodness),
				},
			)
		} else {
			// if we get an error from goodness func, dump it
			logger := app.DecoratedLogger().WithFields(log.Fields{
				"error": err,
				"value": goodness,
			})
			logger.Info("nodegoodness error")
		}
	}
	// Sort goodnesses alphabetically by node address, which are guaranteed
	// to be unique and therefore produce a deterministic ordering. Then
	// topNGoodnesses below can use sort.SliceStable to preserve that ordering
	// in the case of equal goodness values.
	sort.Slice(goodnesses, func(i, j int) bool {
		return goodnesses[i].addr < goodnesses[j].addr
	})
	return goodnesses, goodnessSum
}

func topNGoodnesses(goodnesses []goodnessPair, n int) []goodnessPair {
	// Reverse sort by goodness, preserving original order in case of ties
	sort.SliceStable(goodnesses, func(i, j int) bool {
		return goodnesses[i].goodness > goodnesses[j].goodness
	})
	// pick the top n
	// note: if there is a goodness tie at the boundary, we have to increase
	// n to include all nodes in the tie. To fail to do so would not just
	// be unfair, it would be a determinism bug.
	if n > len(goodnesses) {
		n = len(goodnesses)
	}
	for ; len(goodnesses) > n && goodnesses[n-1].goodness == goodnesses[n].goodness; n++ {
	}
	return goodnesses[:n]
}

// SelectByGoodness deterministically selects one of the active nodes.
//
// If the system var sv.NodeRewardMaxRewarded exists, only that quantity of
// the top nodes by goodness are eligible for rewards.
//
// The specific choice depends on the random number. This random number
// is generated by an external service uniformly in the range [0..uint64_max]
func (app *App) SelectByGoodness(random uint64) (address.Address, error) {
	state := app.GetState().(*backing.State)
	if len(state.Nodes) == 0 {
		return address.Address{}, errors.New("no nodes in nodes list")
	}

	goodnesses, goodnessSum := nodeGoodnesses(app)

	var maxRewarded wkt.Uint64
	err := app.System(sv.NodeRewardMaxRewarded, &maxRewarded)
	// if there's an error, just proceed without filtering
	if err == nil {
		// filter down the top N
		goodnesses = topNGoodnesses(goodnesses, int(maxRewarded))
	} else {
		err = nil
	}

	// goodnesses is a list of tuples. It is currently unordered, as it comes
	// from iterating over a map. We have to sort it for determinism.
	sort.Slice(goodnesses, func(i, j int) bool {
		return goodnesses[i].addr < goodnesses[j].addr
	})

	// bitwise not inverts all bits of 0
	const uint64Max = ^uint64(0)

	// we can't use mod to convert the random value into a selection index,
	// because that would convert a uniform field into a non-uniform field
	// unless by coincidence `uint64_max % random == 0`. Instead, we use
	// non-overflowing muldiv to convert it while preserving uniformity.
	index, err := unsigned.MulDiv(random, goodnessSum, uint64Max)
	if err != nil {
		return address.Address{}, errors.Wrap(err, "overflow generating selection index")
	}

	goodnessSum = 0
	for _, gp := range goodnesses {
		next := goodnessSum + gp.goodness
		if index >= goodnessSum && index < next {
			ra, err := address.Validate(gp.addr)
			if err != nil {
				return address.Address{}, errors.Wrap(err, "winning node had invalid address")
			}
			return ra, nil
		}
		goodnessSum = next
	}
	app.DecoratedLogger().WithFields(log.Fields{
		"index":          index,
		"goodnessSum":    goodnessSum,
		"goodnesses.len": len(goodnesses),
		"method":         "SelectByGoodness",
	}).Warn("failed to select a valid goodness")
	// this happens when the random number is high enough that after multiplication,
	// it exceeds the total goodness of the nodes. In this case, pick the last node.
	gp := goodnesses[len(goodnesses)-1]
	ra, err := address.Validate(gp.addr)
	if err != nil {
		return address.Address{}, errors.Wrap(err, "winning node had invalid address")
	}
	return ra, nil
}
