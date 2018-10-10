package ndau

import (
	"fmt"
	"sort"

	"github.com/oneiro-ndev/chaincode/pkg/chain"
	"github.com/oneiro-ndev/chaincode/pkg/vm"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
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

// IsChaincode is true when the supplied bytes appear to be chaincode
func IsChaincode(code []byte) bool {
	return vm.ConvertToOpcodes(code).IsValid() == nil
}

// BuildVMForTxValidation accepts a transactable and builds a VM that it sets up to call the appropriate
// handler for the given transaction type. All that needs to happen after this is to call Run().
func BuildVMForTxValidation(code []byte, acct backing.AccountData, tx metatx.Transactable,
	signatureSet *bitset256.Bitset256, ts math.Timestamp) (*vm.ChaincodeVM, error) {
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

	// In the context of a transaction, we want the Now opcode to return the transaction's timestamp
	// if it has one, or the current time if it doesn't.
	var nower vm.Nower

	// Similarly, for transactions rand should return values that will be the same across all nodes.
	// We're going to build a seed out of the hash of the SignableBytes of the tx.
	randomer, err := chain.NewSeededRand(tx.SignableBytes())
	if err != nil {
		return nil, err
	}
	nower, err = vm.NewCachingNow(vm.NewTimestamp(ts))
	if err != nil {
		return nil, err
	}
	txID, err := metatx.TxIDOf(tx, TxIDs)
	if err != nil {
		return nil, err
	}
	txIndex := byte(txID)

	bin := buildBinary(code, metatx.NameOf(tx), "")

	switch tx.(type) {
	default:
		// nothing as yet: the point of this switch is to override behaviors
		// for transactions which may require it.
	}
	theVM, err := vm.New(*bin)
	if err != nil {
		return nil, err
	}

	if nower != nil {
		theVM.SetNow(nower)
	}
	if randomer != nil {
		theVM.SetRand(randomer)
	}

	err = theVM.Init(txIndex, acctStruct, txStruct, sigs)
	return theVM, err
}

// BuildVMForTxFees accepts a transactable and builds a VM that it sets up to call the appropriate
// handler for the given transaction type. All that needs to happen after this is to call Run().
func BuildVMForTxFees(code []byte, tx metatx.Transactable, ts math.Timestamp) (*vm.ChaincodeVM, error) {
	txStruct, err := chain.ToValue(tx)
	if err != nil {
		return nil, errors.Wrap(err, "representing tx as chaincode value")
	}

	// In the context of a transaction, we want the Now opcode to return the transaction's timestamp
	// if it has one, or the current time if it doesn't.
	var nower vm.Nower

	// Similarly, for transactions rand should return values that will be the same across all nodes.
	// We're going to build a seed out of the hash of the SignableBytes of the tx.
	randomer, err := chain.NewSeededRand(tx.SignableBytes())
	if err != nil {
		return nil, errors.Wrap(err, "creating seeded rand")
	}
	nower, err = vm.NewCachingNow(vm.NewTimestamp(ts))
	if err != nil {
		return nil, errors.Wrap(err, "creating caching now")
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

	if nower != nil {
		theVM.SetNow(nower)
	}
	if randomer != nil {
		theVM.SetRand(randomer)
	}

	// tx fees are initialized to run the handler associated with the
	// transaction in question, with the length of the full serialized
	// transaction and the transaction struct on the stack
	err = theVM.Init(txIndex, byteLen, txStruct)
	return theVM, errors.Wrap(err, "initializing chaincode vm")
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
) (*vm.ChaincodeVM, error) {
	addrV, err := chain.ToValue(addr)
	if err != nil {
		return nil, err
	}

	acctV, err := chain.ToValue(acct)
	if err != nil {
		return nil, err
	}

	totalAwardV, err := chain.ToValue(totalStake)
	if err != nil {
		return nil, err
	}

	bin := buildBinary(code, fmt.Sprintf("goodness of %s", addr), "")

	theVM, err := vm.New(*bin)
	if err != nil {
		return nil, err
	}

	// In the context of a transaction, we want the Now opcode to return the transaction's timestamp
	// if it has one, or the current time if it doesn't.
	var nower vm.Nower

	nower, err = vm.NewCachingNow(vm.NewTimestamp(ts))
	if err != nil {
		return nil, err
	}
	if nower != nil {
		theVM.SetNow(nower)
	}

	// goodness functions all use the default handler
	err = theVM.Init(0, addrV, acctV, totalAwardV)
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

	// In the context of a transaction, we want the Now opcode to return the transaction's timestamp
	// if it has one, or the current time if it doesn't.
	var nower vm.Nower

	nower, err = vm.NewCachingNow(vm.NewTimestamp(ts))
	if err != nil {
		return nil, errors.Wrap(err, "constructing chaincode nower")
	}
	if nower != nil {
		theVM.SetNow(nower)
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
