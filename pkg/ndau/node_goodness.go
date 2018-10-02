package ndau

import (
	"sort"

	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaumath/pkg/unsigned"
	"github.com/pkg/errors"
)

func (app *App) goodnessOf(addrS string) (int64, error) {
	state := app.GetState().(*backing.State)

	addr, err := address.Validate(addrS)
	if err != nil {
		return 0, err
	}

	acct, hasAcct := state.GetAccount(addr, app.blockTime)
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

	vm, err := BuildVMForNodeGoodness(
		script,
		addr,
		acct,
		math.Ndau(node.TotalStake),
		app.blockTime,
	)
	if err != nil {
		return 0, errors.Wrap(err, "building goodness vm")
	}

	err = vm.Run(false)
	if err != nil {
		return 0, errors.Wrap(err, "running goodness vm")
	}

	goodness, err := vm.Stack().PopAsInt64()
	if err != nil {
		return goodness, errors.Wrap(err, "goodness stack top not numeric")
	}

	return goodness, nil
}

// SelectByGoodness deterministically selects one of the active nodes.
//
// The specific choice depends on the random number. This random number
// is generated by an external service uniformly in the range [0..uint64_max]
func (app *App) SelectByGoodness(random uint64) (address.Address, error) {
	state := app.GetState().(*backing.State)
	if len(state.Nodes) == 0 {
		return address.Address{}, errors.New("no nodes in nodes list")
	}

	type goodnessPair struct {
		addr     string
		goodness uint64
	}
	goodnessSum := uint64(0)
	goodnesses := make([]goodnessPair, 0, len(state.Nodes))
	for addr, node := range state.Nodes {
		if !node.Active {
			continue
		}
		goodness, err := app.goodnessOf(addr)
		if err == nil && goodness > 0 {
			goodnessSum += uint64(goodness)
			goodnesses = append(
				goodnesses,
				goodnessPair{
					addr:     addr,
					goodness: uint64(goodness),
				},
			)
		}
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
	return address.Address{}, errors.New("algorithm error in SelectByGoodness()")
}