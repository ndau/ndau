package ndau

import (
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/oneiro-ndev/chaincode/pkg/chain"
	"github.com/oneiro-ndev/chaincode/pkg/vm"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

func buildBinary(code []byte, name, comment string) *vm.ChasmBinary {
	opcodes := make([]vm.Opcode, len(code))
	for i := 0; i < len(code); i++ {
		opcodes[i] = vm.Opcode(code[i])
	}
	return &vm.ChasmBinary{
		Name:    name,
		Comment: comment,
		Data:    opcodes,
	}
}

// we need to create a distinct seed for this VM which is both deterministic
// and distinct. We get there by xor-ing the block time with the address
func makeSeed(addr address.Address, ts math.Timestamp) []byte {
	addrB := []byte(addr.String())
	tsB := i2b(uint64(ts))
	seed := make([]byte, 64)
	for idx := range seed {
		seed[idx] = addrB[idx%len(addrB)] ^ tsB[idx%len(tsB)]
	}
	return seed
}

func i2b(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func makeSeedInt(a, b, c uint64) []byte {
	ab := i2b(a)
	bb := i2b(b)
	cb := i2b(c)
	seed := make([]byte, 64)
	for idx := range seed {
		seed[idx] = ab[idx%len(ab)] ^ bb[(idx+3)%len(bb)] ^ cb[(idx+6)%len(cb)]
	}
	return seed
}

// we always want to transform ndau VMs in certain predictable ways;
// we bundle those transformations here for convenience
func ndauVM(VM *vm.ChaincodeVM, timestamp math.Timestamp, seed []byte) error {
	// In the context of a transaction, we want the Now opcode to return the transaction's timestamp
	// if it has one, or the current time if it doesn't.
	nower, err := vm.NewCachingNow(vm.NewTimestamp(timestamp))
	if err != nil {
		return err
	}
	VM.SetNow(nower)

	// Similarly, for transactions rand should return values that will be the same across all nodes.
	// We're going to build a seed out of the hash of the SignableBytes of the tx.
	randomer, err := chain.NewSeededRand(seed)
	if err != nil {
		return err
	}
	VM.SetRand(randomer)

	return err
}

// IsChaincode is true when the supplied bytes appear to be chaincode
func IsChaincode(code []byte) bool {
	return vm.ConvertToOpcodes(code).IsValid() == nil
}

// BuildVMForTxValidation accepts a transactable and builds a VM that it sets up to call the appropriate
// handler for the given transaction type. All that needs to happen after this is to call Run().
func BuildVMForTxValidation(
	code []byte,
	acct backing.AccountData,
	tx metatx.Transactable,
	signatureSet *bitset256.Bitset256,
	app *App,
) (*vm.ChaincodeVM, error) {
	acctStruct, err := chain.ToValue(acct)
	if err != nil {
		return nil, err
	}
	txStruct, err := chain.ToValue(tx)
	if err != nil {
		return nil, err
	}
	// because there cannot be very many signatures, we just construct a number corresponding to the
	// lower section of the bitset
	var sigbits int64
	sigmask := int64(1)
	for i := byte(0); i < backing.MaxKeysInAccount; i++ {
		if signatureSet.Get(i) {
			sigbits |= sigmask
		}
		sigmask <<= 1
	}
	sigs := vm.NewNumber(sigbits)

	txID, err := metatx.TxIDOf(tx, TxIDs)
	if err != nil {
		return nil, err
	}
	txIndex := byte(txID)

	bin := buildBinary(code, metatx.NameOf(tx), "")

	args := []vm.Value{acctStruct, txStruct, sigs}

	switch tx.(type) {
	case *ReleaseFromEndowment:
		// RFE transactions need more validation: for the validation scripts
		// we want to run, we need to inject the account data of the destination
		// account onto the bottom of the stack
		destAddr := tx.(*ReleaseFromEndowment).Destination
		dest, _ := app.getAccount(destAddr)
		destStruct, err := chain.ToValue(dest)
		if err != nil {
			return nil, errors.Wrap(err, "creating chain value for dest account data")
		}
		// no prepend builtin, so we're stuck with this ugliness
		args = append([]vm.Value{destStruct}, args...)
	default:
		// nothing as yet: the point of this switch is to override behaviors
		// for transactions which may require it.
	}
	theVM, err := vm.New(*bin)
	if err != nil {
		return nil, err
	}
	err = ndauVM(theVM, app.BlockTime(), tx.SignableBytes())
	if err != nil {
		return nil, err
	}

	err = theVM.Init(txIndex, args...)
	return theVM, err
}

// BuildVMForTxFees accepts a transactable and builds a VM that it sets up to call the appropriate
// handler for the given transaction type. All that needs to happen after this is to call Run().
func BuildVMForTxFees(code []byte, tx metatx.Transactable, ts math.Timestamp) (*vm.ChaincodeVM, error) {
	txStruct, err := chain.ToValue(tx)
	if err != nil {
		return nil, errors.Wrap(err, "representing tx as chaincode value")
	}

	txID, err := metatx.TxIDOf(tx, TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "getting tx id")
	}
	txIndex := byte(txID)

	bin := buildBinary(code, metatx.NameOf(tx), "")

	// in order to get an accourate count of the length of the transaction
	// on the blockchain, we re-serialize it. This should run about as fast
	// as SignableBytes, plus a delta for generating a UUID.
	bytes, err := metatx.Marshal(tx, TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling tx for length")
	}
	byteLen := vm.NewNumber(int64(len(bytes)))

	switch tx.(type) {
	default:
		// nothing as yet: the point of this switch is to override behaviors
		// for transactions which may require it.
	}

	theVM, err := vm.New(*bin)
	if err != nil {
		return nil, errors.Wrap(err, "creating chaincode vm")
	}
	err = ndauVM(theVM, ts, tx.SignableBytes())
	if err != nil {
		return nil, err
	}

	// tx fees are initialized to run the handler associated with the
	// transaction in question, with the length of the full serialized
	// transaction and the transaction struct on the stack
	err = theVM.Init(txIndex, byteLen, txStruct)
	return theVM, errors.Wrap(err, "initializing chaincode vm")
}

// BuildVMForExchangeEAI accepts an exchange account data and builds a VM that it sets up to call
// the default handler.  All that needs to happen after this is to call Run().
func BuildVMForExchangeEAI(code []byte, acct backing.AccountData, sib eai.Rate) (*vm.ChaincodeVM, error) {
	acctV, err := chain.ToValue(acct)
	if err != nil {
		return nil, errors.Wrap(err, "converting account data for exchange EAI chaincode vm")
	}

	sibV, err := chain.ToValue(sib)
	if err != nil {
		return nil, errors.Wrap(err, "converting current SIB rate for chaincode vm")
	}

	bin := buildBinary(code, "Exchange account EAI rate", "")
	theVM, err := vm.New(*bin)
	if err != nil {
		return nil, errors.Wrap(err, "creating exchange EAI chaincode vm")
	}

	err = theVM.Init(0, sibV, acctV)
	if err != nil {
		return nil, errors.Wrap(err, "initializing exchange EAI chaincode vm")
	}

	return theVM, nil
}

// BuildVMForNodeGoodness builds a VM that it sets up to calculate node goodness.
//
// Node goodness functions can currently only use the following three pieces
// of context to make their decision (bottom to top): address, account data,
// total stake.
//
// All that needs to happen after this is to call Run().
func BuildVMForNodeGoodness(
	code []byte,
	addr address.Address,
	acct backing.AccountData,
	totalStake math.Ndau,
	ts math.Timestamp,
	voteStats metast.VoteStats,
) (*vm.ChaincodeVM, error) {
	addrV, err := chain.ToValue(addr)
	if err != nil {
		return nil, errors.Wrap(err, "addr")
	}

	acctV, err := chain.ToValue(acct)
	if err != nil {
		return nil, errors.Wrap(err, "acct")
	}

	totalStakeV, err := chain.ToValue(totalStake)
	if err != nil {
		return nil, errors.Wrap(err, "totalStake")
	}

	var votingHistory []metast.NodeRoundStats
	for _, round := range voteStats.History {
		if nrs, ok := round.Validators[addr.String()]; ok {
			votingHistory = append(votingHistory, nrs)
		}
	}
	votingHistoryV, err := chain.ToValue(votingHistory)
	if err != nil {
		return nil, errors.Wrap(err, "votingHistory")
	}

	bin := buildBinary(code, fmt.Sprintf("goodness of %s", addr), "")

	theVM, err := vm.New(*bin)
	if err != nil {
		return nil, err
	}
	err = ndauVM(theVM, ts, makeSeed(addr, ts))
	if err != nil {
		return nil, err
	}

	// goodness functions all use the default handler
	err = theVM.Init(0, votingHistoryV, addrV, acctV, totalStakeV)
	return theVM, err
}

// BuildVMForNodeDistribution builds a VM that it sets up to calculate distribution
// of node rewards.
//
// All that needs to happen after this is to call Run().
//
// Node distribution functions are expected to return a list of structs on
// top of their stack. These structs must be decorated such that field 10
// is the numeric quantitiy of napu which should be disbursed to that
// costaker.
func BuildVMForNodeDistribution(
	code []byte,
	node address.Address,
	costakers map[string]math.Ndau,
	accounts map[string]backing.AccountData,
	totalAward math.Ndau,
	ts math.Timestamp,
) (*vm.ChaincodeVM, error) {
	// decorate account data with the account's address
	decorateAddr := func(addr string, acct backing.AccountData) (vm.Value, error) {
		acctV, err := chain.ToValue(acct)
		if err != nil {
			return acctV, err
		}
		addrV, err := chain.ToValue(addr)
		if err != nil {
			return acctV, err
		}
		// field 1 is usually Tx_Source, but it also makes sense in a struct
		// context as the address
		acctS, isStruct := acctV.(*vm.Struct)
		if !isStruct {
			return acctV, errors.New("acctV is not a *vm.Struct")
		}
		return acctS.SafeSet(byte(1), addrV)
	}

	nodeV, err := decorateAddr(node.String(), accounts[node.String()])
	if err != nil {
		return nil, errors.Wrap(err, "chaincode value for node")
	}

	totalAwardV, err := chain.ToValue(totalAward)
	if err != nil {
		return nil, errors.Wrap(err, "chaincode value for totalAward")
	}

	costakersV := make([]vm.Value, 0, len(costakers))
	for costaker := range costakers {
		if costaker == node.String() {
			// don't include the node's data among the costakers:
			// nodes get all the leftovers, so we only really care about
			// distribution among everyone else
			continue
		}
		acct, hasAcct := accounts[costaker]
		if !hasAcct {
			continue
		}
		acctV, err := decorateAddr(costaker, acct)
		if err != nil {
			return nil, errors.Wrap(err, "decorating costaker with addr")
		}
		costakersV = append(costakersV, acctV)
	}

	// sort the list of costakers by address for determinism,
	// so we're not thrown off by a node whose distribution script
	// is something like "the first guy in the list gets everything"
	sort.Slice(costakersV, func(i, j int) bool {
		vI, isS := costakersV[i].(*vm.Struct)
		if !isS {
			return false
		}
		aI, err := vI.Get(1)
		if err != nil {
			return false
		}
		vJ, isS := costakersV[j].(*vm.Struct)
		if !isS {
			return false
		}
		aJ, err := vJ.Get(1)
		if err != nil {
			return false
		}
		lt, err := aI.Less(aJ)
		if err != nil {
			return false
		}
		return lt
	})

	bin := buildBinary(code, fmt.Sprintf("distribution for %s", node), "")

	theVM, err := vm.New(*bin)
	if err != nil {
		return nil, errors.Wrap(err, "constructing chaincode vm")
	}
	err = ndauVM(theVM, ts, makeSeed(node, ts))
	if err != nil {
		return nil, err
	}

	// distribution functions all use the default handler
	//
	// stack:
	//  (top) costakers (list of structs of account data decorated with address)
	//        totalAward
	//        node (account data decorated with address)
	err = theVM.Init(0, nodeV, totalAwardV, vm.List(costakersV))
	return theVM, errors.Wrap(err, "initializing chaincode vm")
}

// BuildVMForSIB builds a VM that it sets up to calculate SIB.
//
// The SIB calculation uses exactly two pieces of data: the target price (at
// stack top) and the market price. In principle it doesn't matter what units
// are used for these calculations; any integer pair will do. In practice
// we standardize on Nanocents.
//
// The SIB calculation function returns an integer compatible with eai.Rate:
// the unit is 10^12; 1% is 10^10.
//
// All that needs to happen after this is to call Run().
func BuildVMForSIB(
	code []byte,
	target, market, floor uint64,
	ts math.Timestamp,
) (*vm.ChaincodeVM, error) {
	targetV, err := chain.ToValue(target)
	if err != nil {
		return nil, errors.Wrap(err, "target")
	}
	marketV, err := chain.ToValue(market)
	if err != nil {
		return nil, errors.Wrap(err, "market")
	}
	floorV, err := chain.ToValue(floor)
	if err != nil {
		return nil, errors.Wrap(err, "floor")
	}

	bin := buildBinary(code, "calculate SIB", fmt.Sprintf("market %d; target %d", market, target))

	theVM, err := vm.New(*bin)
	if err != nil {
		return nil, err
	}
	err = ndauVM(theVM, ts, makeSeedInt(target, market, uint64(ts)))
	if err != nil {
		return nil, err
	}

	// goodness functions all use the default handler
	err = theVM.Init(0, floorV, marketV, targetV)
	return theVM, err
}

// BuildVMForRulesValidation builds a VM for rules validation.
//
// Rules validation governs Stake and Unstake transactions. The VM is built
// and run during tx validation to impose additional rules about whether these
// particular transactions are valid. Storing these rules is half the purpose
// of the rules account.
//
// Stack within the VM at init, from top to bottom:
// - total current stake from target staker
// - total current stake from primary staker (0 if no primary stake yet)
// - aggregate stake from primary staker and all costakers (0 if no primary stake)
// - tx
// - target account data
// - stakeTo account data
// - rules account data
//
// Expected output: 0 on top of stack if tx is valid, otherwise non-0
// Additionally, for Unstake: if the second item on the stack is a number
// and non-0, it is interpreted as a math.Duration, added to the block time,
// and the resulting value is used as the expiry date for the hold, which is
// retained. Otherwise, the hold is discarded immediately.
func BuildVMForRulesValidation(
	tx metatx.Transactable,
	state *backing.State,
	rulesAcct ...address.Address,
) (*vm.ChaincodeVM, error) {
	id, err := metatx.TxIDOf(tx, TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "tx id")
	}

	var target, stakeTo, rules, primary address.Address
	switch t := tx.(type) {
	case *Stake:
		target = t.Target
		stakeTo = t.StakeTo
		rules = t.Rules
		if stakeTo == rules {
			primary = target
		} else {
			primary = stakeTo
		}
	case *Unstake:
		target = t.Target
		stakeTo = t.StakeTo
		rules = t.Rules
		if stakeTo == rules {
			primary = target
		} else {
			primary = stakeTo
		}
	case *ResolveStake:
		target = t.Target
		rules = t.Rules
		primary = t.Target
	case *RegisterNode:
		if len(rulesAcct) != 1 {
			return nil, fmt.Errorf("expect 1 rules account; got %d", len(rulesAcct))
		}
		target = t.Node
		primary = t.Node
		rules = rulesAcct[0]
	default:
		return nil, fmt.Errorf("Rules Validation VM should not be constructed for %T", tx)
	}

	var aggregateQ, primaryTotalQ, targetTotalQ math.Ndau

	targetTotalQ = state.TotalStake(target, primary, rules)
	targetTotalV, err := chain.ToValue(targetTotalQ)
	if err != nil {
		return nil, errors.Wrap(err, "target total")
	}

	primaryTotalQ = state.TotalStake(primary, primary, rules)
	primaryTotalV, err := chain.ToValue(primaryTotalQ)
	if err != nil {
		return nil, errors.Wrap(err, "primary total")
	}

	aggregate := state.AggregateStake(primary, rules)
	if aggregate != nil {
		aggregateQ = aggregate.Qty
	}
	aggregateV, err := chain.ToValue(aggregateQ)
	if err != nil {
		return nil, errors.Wrap(err, "aggregate from primary")
	}

	txV, err := chain.ToValue(tx)
	if err != nil {
		return nil, errors.Wrap(err, "tx")
	}
	targetV, err := chain.ToValue(target)
	if err != nil {
		return nil, errors.Wrap(err, "target")
	}
	stakeToV, err := chain.ToValue(stakeTo)
	if err != nil {
		return nil, errors.Wrap(err, "stakeTo")
	}
	rulesV, err := chain.ToValue(rules)
	if err != nil {
		return nil, errors.Wrap(err, "rules")
	}
	primaryV, err := chain.ToValue(primary)
	if err != nil {
		return nil, errors.Wrap(err, "primary")
	}

	rulesAcctData, ok := state.Accounts[rules.String()]
	if !ok {
		return nil, errors.New("rules account does not exist")
	}
	if rulesAcctData.StakeRules == nil {
		return nil, errors.New("rules account has no stake rules")
	}

	bin := buildBinary(rulesAcctData.StakeRules.Script, "Rules validation", "")
	theVM, err := vm.New(*bin)
	if err != nil {
		return nil, errors.Wrap(err, "creating vm")
	}

	err = theVM.Init(
		byte(id),
		primaryV, rulesV, stakeToV, targetV,
		txV,
		aggregateV, primaryTotalV, targetTotalV,
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing vm")
	}

	return theVM, nil
}
