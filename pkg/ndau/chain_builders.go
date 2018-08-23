package ndau

import (
	"errors"

	"github.com/oneiro-ndev/chaincode/pkg/chain"
	"github.com/oneiro-ndev/chaincode/pkg/vm"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
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

	var bin *vm.ChasmBinary

	txIndex := byte(0)
	switch tx.(type) {
	case *Transfer:
		bin = buildBinary(code, "Transfer", "")
		txIndex, err = GetID(tx)
		if err != nil {
			return nil, err
		}
		nower, _ = vm.NewCachingNow(vm.NewTimestamp(ts))
	default:
		return nil, errors.New("unhandled transactable type")
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
