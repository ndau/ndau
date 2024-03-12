// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package minter

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// MinterABI is the input ABI used to generate the binding from.
const MinterABI = "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_validators\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"_tokenAddress\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"age\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"bounty\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"burns\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentBlock\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"delegate\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"delegation\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNo\",\"type\":\"uint256\"},{\"internalType\":\"address[]\",\"name\":\"candidates\",\"type\":\"address[]\"}],\"name\":\"elect\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBurns\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"txHash\",\"type\":\"bytes32\"}],\"name\":\"getMinted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"txHash\",\"type\":\"bytes32\"}],\"name\":\"getProofVote\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"proposal\",\"type\":\"bytes32\"}],\"name\":\"getProofVoters\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"structHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"getSigner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTallies\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNo\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"proposalIndex\",\"type\":\"uint256\"}],\"name\":\"getTallyCandidates\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNo\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"proposal\",\"type\":\"bytes32\"}],\"name\":\"getTallyProposalVote\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNo\",\"type\":\"uint256\"}],\"name\":\"getTallyProposals\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNo\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"voter\",\"type\":\"address\"}],\"name\":\"getTallyVoterBallot\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNo\",\"type\":\"uint256\"}],\"name\":\"getTallyVoters\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getValidators\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"mode\",\"type\":\"uint256\"}],\"name\":\"getVotingPower\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"isThreshold1\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"isThreshold2\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minter\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"proof\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"votes\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"minted\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"stake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"tallies\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"tally\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"votes\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"called\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"threholdFine\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"threshold1\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"threshold2\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tokenAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"validator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"validators\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"tx_hash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"block_no\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"verifyMintingSigner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"txHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"blockNo\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"vote\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"txHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"blockNo\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"vote\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// Minter is an auto generated Go binding around an Ethereum contract.
type Minter struct {
	MinterCaller     // Read-only binding to the contract
	MinterTransactor // Write-only binding to the contract
	MinterFilterer   // Log filterer for contract events
}

// MinterCaller is an auto generated read-only Go binding around an Ethereum contract.
type MinterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MinterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MinterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MinterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MinterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MinterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MinterSession struct {
	Contract     *Minter           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MinterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MinterCallerSession struct {
	Contract *MinterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// MinterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MinterTransactorSession struct {
	Contract     *MinterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MinterRaw is an auto generated low-level Go binding around an Ethereum contract.
type MinterRaw struct {
	Contract *Minter // Generic contract binding to access the raw methods on
}

// MinterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MinterCallerRaw struct {
	Contract *MinterCaller // Generic read-only contract binding to access the raw methods on
}

// MinterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MinterTransactorRaw struct {
	Contract *MinterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMinter creates a new instance of Minter, bound to a specific deployed contract.
func NewMinter(address common.Address, backend bind.ContractBackend) (*Minter, error) {
	contract, err := bindMinter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Minter{MinterCaller: MinterCaller{contract: contract}, MinterTransactor: MinterTransactor{contract: contract}, MinterFilterer: MinterFilterer{contract: contract}}, nil
}

// NewMinterCaller creates a new read-only instance of Minter, bound to a specific deployed contract.
func NewMinterCaller(address common.Address, caller bind.ContractCaller) (*MinterCaller, error) {
	contract, err := bindMinter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MinterCaller{contract: contract}, nil
}

// NewMinterTransactor creates a new write-only instance of Minter, bound to a specific deployed contract.
func NewMinterTransactor(address common.Address, transactor bind.ContractTransactor) (*MinterTransactor, error) {
	contract, err := bindMinter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MinterTransactor{contract: contract}, nil
}

// NewMinterFilterer creates a new log filterer instance of Minter, bound to a specific deployed contract.
func NewMinterFilterer(address common.Address, filterer bind.ContractFilterer) (*MinterFilterer, error) {
	contract, err := bindMinter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MinterFilterer{contract: contract}, nil
}

// bindMinter binds a generic wrapper to an already deployed contract.
func bindMinter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MinterABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Minter *MinterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Minter.Contract.MinterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Minter *MinterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Minter.Contract.MinterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Minter *MinterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Minter.Contract.MinterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Minter *MinterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Minter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Minter *MinterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Minter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Minter *MinterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Minter.Contract.contract.Transact(opts, method, params...)
}

// Age is a free data retrieval call binding the contract method 0xd7efb7b3.
//
// Solidity: function age(bytes32 ) view returns(uint256)
func (_Minter *MinterCaller) Age(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "age", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Age is a free data retrieval call binding the contract method 0xd7efb7b3.
//
// Solidity: function age(bytes32 ) view returns(uint256)
func (_Minter *MinterSession) Age(arg0 [32]byte) (*big.Int, error) {
	return _Minter.Contract.Age(&_Minter.CallOpts, arg0)
}

// Age is a free data retrieval call binding the contract method 0xd7efb7b3.
//
// Solidity: function age(bytes32 ) view returns(uint256)
func (_Minter *MinterCallerSession) Age(arg0 [32]byte) (*big.Int, error) {
	return _Minter.Contract.Age(&_Minter.CallOpts, arg0)
}

// Bounty is a free data retrieval call binding the contract method 0x19f8c885.
//
// Solidity: function bounty(address ) view returns(uint256)
func (_Minter *MinterCaller) Bounty(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "bounty", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Bounty is a free data retrieval call binding the contract method 0x19f8c885.
//
// Solidity: function bounty(address ) view returns(uint256)
func (_Minter *MinterSession) Bounty(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Bounty(&_Minter.CallOpts, arg0)
}

// Bounty is a free data retrieval call binding the contract method 0x19f8c885.
//
// Solidity: function bounty(address ) view returns(uint256)
func (_Minter *MinterCallerSession) Bounty(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Bounty(&_Minter.CallOpts, arg0)
}

// Burns is a free data retrieval call binding the contract method 0xa86eb292.
//
// Solidity: function burns(uint256 ) view returns(bytes32)
func (_Minter *MinterCaller) Burns(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "burns", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Burns is a free data retrieval call binding the contract method 0xa86eb292.
//
// Solidity: function burns(uint256 ) view returns(bytes32)
func (_Minter *MinterSession) Burns(arg0 *big.Int) ([32]byte, error) {
	return _Minter.Contract.Burns(&_Minter.CallOpts, arg0)
}

// Burns is a free data retrieval call binding the contract method 0xa86eb292.
//
// Solidity: function burns(uint256 ) view returns(bytes32)
func (_Minter *MinterCallerSession) Burns(arg0 *big.Int) ([32]byte, error) {
	return _Minter.Contract.Burns(&_Minter.CallOpts, arg0)
}

// CurrentBlock is a free data retrieval call binding the contract method 0xe12ed13c.
//
// Solidity: function currentBlock() view returns(uint256)
func (_Minter *MinterCaller) CurrentBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "currentBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CurrentBlock is a free data retrieval call binding the contract method 0xe12ed13c.
//
// Solidity: function currentBlock() view returns(uint256)
func (_Minter *MinterSession) CurrentBlock() (*big.Int, error) {
	return _Minter.Contract.CurrentBlock(&_Minter.CallOpts)
}

// CurrentBlock is a free data retrieval call binding the contract method 0xe12ed13c.
//
// Solidity: function currentBlock() view returns(uint256)
func (_Minter *MinterCallerSession) CurrentBlock() (*big.Int, error) {
	return _Minter.Contract.CurrentBlock(&_Minter.CallOpts)
}

// Delegation is a free data retrieval call binding the contract method 0xeed50a32.
//
// Solidity: function delegation(address ) view returns(uint256)
func (_Minter *MinterCaller) Delegation(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "delegation", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Delegation is a free data retrieval call binding the contract method 0xeed50a32.
//
// Solidity: function delegation(address ) view returns(uint256)
func (_Minter *MinterSession) Delegation(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Delegation(&_Minter.CallOpts, arg0)
}

// Delegation is a free data retrieval call binding the contract method 0xeed50a32.
//
// Solidity: function delegation(address ) view returns(uint256)
func (_Minter *MinterCallerSession) Delegation(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Delegation(&_Minter.CallOpts, arg0)
}

// GetBurns is a free data retrieval call binding the contract method 0x88b228fc.
//
// Solidity: function getBurns() view returns(bytes32[])
func (_Minter *MinterCaller) GetBurns(opts *bind.CallOpts) ([][32]byte, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getBurns")

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetBurns is a free data retrieval call binding the contract method 0x88b228fc.
//
// Solidity: function getBurns() view returns(bytes32[])
func (_Minter *MinterSession) GetBurns() ([][32]byte, error) {
	return _Minter.Contract.GetBurns(&_Minter.CallOpts)
}

// GetBurns is a free data retrieval call binding the contract method 0x88b228fc.
//
// Solidity: function getBurns() view returns(bytes32[])
func (_Minter *MinterCallerSession) GetBurns() ([][32]byte, error) {
	return _Minter.Contract.GetBurns(&_Minter.CallOpts)
}

// GetMinted is a free data retrieval call binding the contract method 0xa2be6bd2.
//
// Solidity: function getMinted(bytes32 txHash) view returns(bool)
func (_Minter *MinterCaller) GetMinted(opts *bind.CallOpts, txHash [32]byte) (bool, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getMinted", txHash)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetMinted is a free data retrieval call binding the contract method 0xa2be6bd2.
//
// Solidity: function getMinted(bytes32 txHash) view returns(bool)
func (_Minter *MinterSession) GetMinted(txHash [32]byte) (bool, error) {
	return _Minter.Contract.GetMinted(&_Minter.CallOpts, txHash)
}

// GetMinted is a free data retrieval call binding the contract method 0xa2be6bd2.
//
// Solidity: function getMinted(bytes32 txHash) view returns(bool)
func (_Minter *MinterCallerSession) GetMinted(txHash [32]byte) (bool, error) {
	return _Minter.Contract.GetMinted(&_Minter.CallOpts, txHash)
}

// GetProofVote is a free data retrieval call binding the contract method 0xd3dd0f87.
//
// Solidity: function getProofVote(bytes32 txHash) view returns(uint256)
func (_Minter *MinterCaller) GetProofVote(opts *bind.CallOpts, txHash [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getProofVote", txHash)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProofVote is a free data retrieval call binding the contract method 0xd3dd0f87.
//
// Solidity: function getProofVote(bytes32 txHash) view returns(uint256)
func (_Minter *MinterSession) GetProofVote(txHash [32]byte) (*big.Int, error) {
	return _Minter.Contract.GetProofVote(&_Minter.CallOpts, txHash)
}

// GetProofVote is a free data retrieval call binding the contract method 0xd3dd0f87.
//
// Solidity: function getProofVote(bytes32 txHash) view returns(uint256)
func (_Minter *MinterCallerSession) GetProofVote(txHash [32]byte) (*big.Int, error) {
	return _Minter.Contract.GetProofVote(&_Minter.CallOpts, txHash)
}

// GetProofVoters is a free data retrieval call binding the contract method 0xfde3ffb2.
//
// Solidity: function getProofVoters(bytes32 proposal) view returns(address[])
func (_Minter *MinterCaller) GetProofVoters(opts *bind.CallOpts, proposal [32]byte) ([]common.Address, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getProofVoters", proposal)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetProofVoters is a free data retrieval call binding the contract method 0xfde3ffb2.
//
// Solidity: function getProofVoters(bytes32 proposal) view returns(address[])
func (_Minter *MinterSession) GetProofVoters(proposal [32]byte) ([]common.Address, error) {
	return _Minter.Contract.GetProofVoters(&_Minter.CallOpts, proposal)
}

// GetProofVoters is a free data retrieval call binding the contract method 0xfde3ffb2.
//
// Solidity: function getProofVoters(bytes32 proposal) view returns(address[])
func (_Minter *MinterCallerSession) GetProofVoters(proposal [32]byte) ([]common.Address, error) {
	return _Minter.Contract.GetProofVoters(&_Minter.CallOpts, proposal)
}

// GetSigner is a free data retrieval call binding the contract method 0xf96ddf7a.
//
// Solidity: function getSigner(bytes32 structHash, uint8 v, bytes32 r, bytes32 s) view returns(address)
func (_Minter *MinterCaller) GetSigner(opts *bind.CallOpts, structHash [32]byte, v uint8, r [32]byte, s [32]byte) (common.Address, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getSigner", structHash, v, r, s)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetSigner is a free data retrieval call binding the contract method 0xf96ddf7a.
//
// Solidity: function getSigner(bytes32 structHash, uint8 v, bytes32 r, bytes32 s) view returns(address)
func (_Minter *MinterSession) GetSigner(structHash [32]byte, v uint8, r [32]byte, s [32]byte) (common.Address, error) {
	return _Minter.Contract.GetSigner(&_Minter.CallOpts, structHash, v, r, s)
}

// GetSigner is a free data retrieval call binding the contract method 0xf96ddf7a.
//
// Solidity: function getSigner(bytes32 structHash, uint8 v, bytes32 r, bytes32 s) view returns(address)
func (_Minter *MinterCallerSession) GetSigner(structHash [32]byte, v uint8, r [32]byte, s [32]byte) (common.Address, error) {
	return _Minter.Contract.GetSigner(&_Minter.CallOpts, structHash, v, r, s)
}

// GetTallies is a free data retrieval call binding the contract method 0x8e1c690a.
//
// Solidity: function getTallies() view returns(uint256[])
func (_Minter *MinterCaller) GetTallies(opts *bind.CallOpts) ([]*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getTallies")

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetTallies is a free data retrieval call binding the contract method 0x8e1c690a.
//
// Solidity: function getTallies() view returns(uint256[])
func (_Minter *MinterSession) GetTallies() ([]*big.Int, error) {
	return _Minter.Contract.GetTallies(&_Minter.CallOpts)
}

// GetTallies is a free data retrieval call binding the contract method 0x8e1c690a.
//
// Solidity: function getTallies() view returns(uint256[])
func (_Minter *MinterCallerSession) GetTallies() ([]*big.Int, error) {
	return _Minter.Contract.GetTallies(&_Minter.CallOpts)
}

// GetTallyCandidates is a free data retrieval call binding the contract method 0x375b7314.
//
// Solidity: function getTallyCandidates(uint256 blockNo, uint256 proposalIndex) view returns(address[])
func (_Minter *MinterCaller) GetTallyCandidates(opts *bind.CallOpts, blockNo *big.Int, proposalIndex *big.Int) ([]common.Address, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getTallyCandidates", blockNo, proposalIndex)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetTallyCandidates is a free data retrieval call binding the contract method 0x375b7314.
//
// Solidity: function getTallyCandidates(uint256 blockNo, uint256 proposalIndex) view returns(address[])
func (_Minter *MinterSession) GetTallyCandidates(blockNo *big.Int, proposalIndex *big.Int) ([]common.Address, error) {
	return _Minter.Contract.GetTallyCandidates(&_Minter.CallOpts, blockNo, proposalIndex)
}

// GetTallyCandidates is a free data retrieval call binding the contract method 0x375b7314.
//
// Solidity: function getTallyCandidates(uint256 blockNo, uint256 proposalIndex) view returns(address[])
func (_Minter *MinterCallerSession) GetTallyCandidates(blockNo *big.Int, proposalIndex *big.Int) ([]common.Address, error) {
	return _Minter.Contract.GetTallyCandidates(&_Minter.CallOpts, blockNo, proposalIndex)
}

// GetTallyProposalVote is a free data retrieval call binding the contract method 0x2e1f7d21.
//
// Solidity: function getTallyProposalVote(uint256 blockNo, bytes32 proposal) view returns(uint256)
func (_Minter *MinterCaller) GetTallyProposalVote(opts *bind.CallOpts, blockNo *big.Int, proposal [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getTallyProposalVote", blockNo, proposal)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTallyProposalVote is a free data retrieval call binding the contract method 0x2e1f7d21.
//
// Solidity: function getTallyProposalVote(uint256 blockNo, bytes32 proposal) view returns(uint256)
func (_Minter *MinterSession) GetTallyProposalVote(blockNo *big.Int, proposal [32]byte) (*big.Int, error) {
	return _Minter.Contract.GetTallyProposalVote(&_Minter.CallOpts, blockNo, proposal)
}

// GetTallyProposalVote is a free data retrieval call binding the contract method 0x2e1f7d21.
//
// Solidity: function getTallyProposalVote(uint256 blockNo, bytes32 proposal) view returns(uint256)
func (_Minter *MinterCallerSession) GetTallyProposalVote(blockNo *big.Int, proposal [32]byte) (*big.Int, error) {
	return _Minter.Contract.GetTallyProposalVote(&_Minter.CallOpts, blockNo, proposal)
}

// GetTallyProposals is a free data retrieval call binding the contract method 0x80601667.
//
// Solidity: function getTallyProposals(uint256 blockNo) view returns(bytes32[])
func (_Minter *MinterCaller) GetTallyProposals(opts *bind.CallOpts, blockNo *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getTallyProposals", blockNo)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetTallyProposals is a free data retrieval call binding the contract method 0x80601667.
//
// Solidity: function getTallyProposals(uint256 blockNo) view returns(bytes32[])
func (_Minter *MinterSession) GetTallyProposals(blockNo *big.Int) ([][32]byte, error) {
	return _Minter.Contract.GetTallyProposals(&_Minter.CallOpts, blockNo)
}

// GetTallyProposals is a free data retrieval call binding the contract method 0x80601667.
//
// Solidity: function getTallyProposals(uint256 blockNo) view returns(bytes32[])
func (_Minter *MinterCallerSession) GetTallyProposals(blockNo *big.Int) ([][32]byte, error) {
	return _Minter.Contract.GetTallyProposals(&_Minter.CallOpts, blockNo)
}

// GetTallyVoterBallot is a free data retrieval call binding the contract method 0x009a4941.
//
// Solidity: function getTallyVoterBallot(uint256 blockNo, address voter) view returns(bytes32)
func (_Minter *MinterCaller) GetTallyVoterBallot(opts *bind.CallOpts, blockNo *big.Int, voter common.Address) ([32]byte, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getTallyVoterBallot", blockNo, voter)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetTallyVoterBallot is a free data retrieval call binding the contract method 0x009a4941.
//
// Solidity: function getTallyVoterBallot(uint256 blockNo, address voter) view returns(bytes32)
func (_Minter *MinterSession) GetTallyVoterBallot(blockNo *big.Int, voter common.Address) ([32]byte, error) {
	return _Minter.Contract.GetTallyVoterBallot(&_Minter.CallOpts, blockNo, voter)
}

// GetTallyVoterBallot is a free data retrieval call binding the contract method 0x009a4941.
//
// Solidity: function getTallyVoterBallot(uint256 blockNo, address voter) view returns(bytes32)
func (_Minter *MinterCallerSession) GetTallyVoterBallot(blockNo *big.Int, voter common.Address) ([32]byte, error) {
	return _Minter.Contract.GetTallyVoterBallot(&_Minter.CallOpts, blockNo, voter)
}

// GetTallyVoters is a free data retrieval call binding the contract method 0x4b341bef.
//
// Solidity: function getTallyVoters(uint256 blockNo) view returns(address[])
func (_Minter *MinterCaller) GetTallyVoters(opts *bind.CallOpts, blockNo *big.Int) ([]common.Address, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getTallyVoters", blockNo)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetTallyVoters is a free data retrieval call binding the contract method 0x4b341bef.
//
// Solidity: function getTallyVoters(uint256 blockNo) view returns(address[])
func (_Minter *MinterSession) GetTallyVoters(blockNo *big.Int) ([]common.Address, error) {
	return _Minter.Contract.GetTallyVoters(&_Minter.CallOpts, blockNo)
}

// GetTallyVoters is a free data retrieval call binding the contract method 0x4b341bef.
//
// Solidity: function getTallyVoters(uint256 blockNo) view returns(address[])
func (_Minter *MinterCallerSession) GetTallyVoters(blockNo *big.Int) ([]common.Address, error) {
	return _Minter.Contract.GetTallyVoters(&_Minter.CallOpts, blockNo)
}

// GetValidators is a free data retrieval call binding the contract method 0xb7ab4db5.
//
// Solidity: function getValidators() view returns(address[])
func (_Minter *MinterCaller) GetValidators(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getValidators")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetValidators is a free data retrieval call binding the contract method 0xb7ab4db5.
//
// Solidity: function getValidators() view returns(address[])
func (_Minter *MinterSession) GetValidators() ([]common.Address, error) {
	return _Minter.Contract.GetValidators(&_Minter.CallOpts)
}

// GetValidators is a free data retrieval call binding the contract method 0xb7ab4db5.
//
// Solidity: function getValidators() view returns(address[])
func (_Minter *MinterCallerSession) GetValidators() ([]common.Address, error) {
	return _Minter.Contract.GetValidators(&_Minter.CallOpts)
}

// GetVotingPower is a free data retrieval call binding the contract method 0x9c6d2976.
//
// Solidity: function getVotingPower(uint256 mode) view returns(uint256)
func (_Minter *MinterCaller) GetVotingPower(opts *bind.CallOpts, mode *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "getVotingPower", mode)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetVotingPower is a free data retrieval call binding the contract method 0x9c6d2976.
//
// Solidity: function getVotingPower(uint256 mode) view returns(uint256)
func (_Minter *MinterSession) GetVotingPower(mode *big.Int) (*big.Int, error) {
	return _Minter.Contract.GetVotingPower(&_Minter.CallOpts, mode)
}

// GetVotingPower is a free data retrieval call binding the contract method 0x9c6d2976.
//
// Solidity: function getVotingPower(uint256 mode) view returns(uint256)
func (_Minter *MinterCallerSession) GetVotingPower(mode *big.Int) (*big.Int, error) {
	return _Minter.Contract.GetVotingPower(&_Minter.CallOpts, mode)
}

// IsThreshold1 is a free data retrieval call binding the contract method 0x16c7cbd1.
//
// Solidity: function isThreshold1(address _address) view returns(bool)
func (_Minter *MinterCaller) IsThreshold1(opts *bind.CallOpts, _address common.Address) (bool, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "isThreshold1", _address)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsThreshold1 is a free data retrieval call binding the contract method 0x16c7cbd1.
//
// Solidity: function isThreshold1(address _address) view returns(bool)
func (_Minter *MinterSession) IsThreshold1(_address common.Address) (bool, error) {
	return _Minter.Contract.IsThreshold1(&_Minter.CallOpts, _address)
}

// IsThreshold1 is a free data retrieval call binding the contract method 0x16c7cbd1.
//
// Solidity: function isThreshold1(address _address) view returns(bool)
func (_Minter *MinterCallerSession) IsThreshold1(_address common.Address) (bool, error) {
	return _Minter.Contract.IsThreshold1(&_Minter.CallOpts, _address)
}

// IsThreshold2 is a free data retrieval call binding the contract method 0x11bec787.
//
// Solidity: function isThreshold2(address _address) view returns(bool)
func (_Minter *MinterCaller) IsThreshold2(opts *bind.CallOpts, _address common.Address) (bool, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "isThreshold2", _address)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsThreshold2 is a free data retrieval call binding the contract method 0x11bec787.
//
// Solidity: function isThreshold2(address _address) view returns(bool)
func (_Minter *MinterSession) IsThreshold2(_address common.Address) (bool, error) {
	return _Minter.Contract.IsThreshold2(&_Minter.CallOpts, _address)
}

// IsThreshold2 is a free data retrieval call binding the contract method 0x11bec787.
//
// Solidity: function isThreshold2(address _address) view returns(bool)
func (_Minter *MinterCallerSession) IsThreshold2(_address common.Address) (bool, error) {
	return _Minter.Contract.IsThreshold2(&_Minter.CallOpts, _address)
}

// Minter is a free data retrieval call binding the contract method 0x07546172.
//
// Solidity: function minter() view returns(address)
func (_Minter *MinterCaller) Minter(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "minter")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Minter is a free data retrieval call binding the contract method 0x07546172.
//
// Solidity: function minter() view returns(address)
func (_Minter *MinterSession) Minter() (common.Address, error) {
	return _Minter.Contract.Minter(&_Minter.CallOpts)
}

// Minter is a free data retrieval call binding the contract method 0x07546172.
//
// Solidity: function minter() view returns(address)
func (_Minter *MinterCallerSession) Minter() (common.Address, error) {
	return _Minter.Contract.Minter(&_Minter.CallOpts)
}

// Proof is a free data retrieval call binding the contract method 0xbcfb013f.
//
// Solidity: function proof(bytes32 ) view returns(uint256 votes, bool minted)
func (_Minter *MinterCaller) Proof(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Votes  *big.Int
	Minted bool
}, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "proof", arg0)

	outstruct := new(struct {
		Votes  *big.Int
		Minted bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Votes = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Minted = *abi.ConvertType(out[1], new(bool)).(*bool)

	return *outstruct, err

}

// Proof is a free data retrieval call binding the contract method 0xbcfb013f.
//
// Solidity: function proof(bytes32 ) view returns(uint256 votes, bool minted)
func (_Minter *MinterSession) Proof(arg0 [32]byte) (struct {
	Votes  *big.Int
	Minted bool
}, error) {
	return _Minter.Contract.Proof(&_Minter.CallOpts, arg0)
}

// Proof is a free data retrieval call binding the contract method 0xbcfb013f.
//
// Solidity: function proof(bytes32 ) view returns(uint256 votes, bool minted)
func (_Minter *MinterCallerSession) Proof(arg0 [32]byte) (struct {
	Votes  *big.Int
	Minted bool
}, error) {
	return _Minter.Contract.Proof(&_Minter.CallOpts, arg0)
}

// Stake is a free data retrieval call binding the contract method 0x26476204.
//
// Solidity: function stake(address ) view returns(uint256)
func (_Minter *MinterCaller) Stake(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "stake", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Stake is a free data retrieval call binding the contract method 0x26476204.
//
// Solidity: function stake(address ) view returns(uint256)
func (_Minter *MinterSession) Stake(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Stake(&_Minter.CallOpts, arg0)
}

// Stake is a free data retrieval call binding the contract method 0x26476204.
//
// Solidity: function stake(address ) view returns(uint256)
func (_Minter *MinterCallerSession) Stake(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Stake(&_Minter.CallOpts, arg0)
}

// Tallies is a free data retrieval call binding the contract method 0x1a32b237.
//
// Solidity: function tallies(uint256 ) view returns(uint256)
func (_Minter *MinterCaller) Tallies(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "tallies", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Tallies is a free data retrieval call binding the contract method 0x1a32b237.
//
// Solidity: function tallies(uint256 ) view returns(uint256)
func (_Minter *MinterSession) Tallies(arg0 *big.Int) (*big.Int, error) {
	return _Minter.Contract.Tallies(&_Minter.CallOpts, arg0)
}

// Tallies is a free data retrieval call binding the contract method 0x1a32b237.
//
// Solidity: function tallies(uint256 ) view returns(uint256)
func (_Minter *MinterCallerSession) Tallies(arg0 *big.Int) (*big.Int, error) {
	return _Minter.Contract.Tallies(&_Minter.CallOpts, arg0)
}

// Tally is a free data retrieval call binding the contract method 0xed8b6b31.
//
// Solidity: function tally(uint256 ) view returns(uint256 votes, bool called)
func (_Minter *MinterCaller) Tally(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Votes  *big.Int
	Called bool
}, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "tally", arg0)

	outstruct := new(struct {
		Votes  *big.Int
		Called bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Votes = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Called = *abi.ConvertType(out[1], new(bool)).(*bool)

	return *outstruct, err

}

// Tally is a free data retrieval call binding the contract method 0xed8b6b31.
//
// Solidity: function tally(uint256 ) view returns(uint256 votes, bool called)
func (_Minter *MinterSession) Tally(arg0 *big.Int) (struct {
	Votes  *big.Int
	Called bool
}, error) {
	return _Minter.Contract.Tally(&_Minter.CallOpts, arg0)
}

// Tally is a free data retrieval call binding the contract method 0xed8b6b31.
//
// Solidity: function tally(uint256 ) view returns(uint256 votes, bool called)
func (_Minter *MinterCallerSession) Tally(arg0 *big.Int) (struct {
	Votes  *big.Int
	Called bool
}, error) {
	return _Minter.Contract.Tally(&_Minter.CallOpts, arg0)
}

// ThreholdFine is a free data retrieval call binding the contract method 0xc1e30473.
//
// Solidity: function threholdFine() view returns(uint256)
func (_Minter *MinterCaller) ThreholdFine(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "threholdFine")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ThreholdFine is a free data retrieval call binding the contract method 0xc1e30473.
//
// Solidity: function threholdFine() view returns(uint256)
func (_Minter *MinterSession) ThreholdFine() (*big.Int, error) {
	return _Minter.Contract.ThreholdFine(&_Minter.CallOpts)
}

// ThreholdFine is a free data retrieval call binding the contract method 0xc1e30473.
//
// Solidity: function threholdFine() view returns(uint256)
func (_Minter *MinterCallerSession) ThreholdFine() (*big.Int, error) {
	return _Minter.Contract.ThreholdFine(&_Minter.CallOpts)
}

// Threshold1 is a free data retrieval call binding the contract method 0x49dfb4d8.
//
// Solidity: function threshold1(address ) view returns(uint256)
func (_Minter *MinterCaller) Threshold1(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "threshold1", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Threshold1 is a free data retrieval call binding the contract method 0x49dfb4d8.
//
// Solidity: function threshold1(address ) view returns(uint256)
func (_Minter *MinterSession) Threshold1(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Threshold1(&_Minter.CallOpts, arg0)
}

// Threshold1 is a free data retrieval call binding the contract method 0x49dfb4d8.
//
// Solidity: function threshold1(address ) view returns(uint256)
func (_Minter *MinterCallerSession) Threshold1(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Threshold1(&_Minter.CallOpts, arg0)
}

// Threshold2 is a free data retrieval call binding the contract method 0x1f025f58.
//
// Solidity: function threshold2(address ) view returns(uint256)
func (_Minter *MinterCaller) Threshold2(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "threshold2", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Threshold2 is a free data retrieval call binding the contract method 0x1f025f58.
//
// Solidity: function threshold2(address ) view returns(uint256)
func (_Minter *MinterSession) Threshold2(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Threshold2(&_Minter.CallOpts, arg0)
}

// Threshold2 is a free data retrieval call binding the contract method 0x1f025f58.
//
// Solidity: function threshold2(address ) view returns(uint256)
func (_Minter *MinterCallerSession) Threshold2(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Threshold2(&_Minter.CallOpts, arg0)
}

// TokenAddress is a free data retrieval call binding the contract method 0x9d76ea58.
//
// Solidity: function tokenAddress() view returns(address)
func (_Minter *MinterCaller) TokenAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "tokenAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenAddress is a free data retrieval call binding the contract method 0x9d76ea58.
//
// Solidity: function tokenAddress() view returns(address)
func (_Minter *MinterSession) TokenAddress() (common.Address, error) {
	return _Minter.Contract.TokenAddress(&_Minter.CallOpts)
}

// TokenAddress is a free data retrieval call binding the contract method 0x9d76ea58.
//
// Solidity: function tokenAddress() view returns(address)
func (_Minter *MinterCallerSession) TokenAddress() (common.Address, error) {
	return _Minter.Contract.TokenAddress(&_Minter.CallOpts)
}

// Validator is a free data retrieval call binding the contract method 0x223b3b7a.
//
// Solidity: function validator(address ) view returns(uint256)
func (_Minter *MinterCaller) Validator(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "validator", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Validator is a free data retrieval call binding the contract method 0x223b3b7a.
//
// Solidity: function validator(address ) view returns(uint256)
func (_Minter *MinterSession) Validator(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Validator(&_Minter.CallOpts, arg0)
}

// Validator is a free data retrieval call binding the contract method 0x223b3b7a.
//
// Solidity: function validator(address ) view returns(uint256)
func (_Minter *MinterCallerSession) Validator(arg0 common.Address) (*big.Int, error) {
	return _Minter.Contract.Validator(&_Minter.CallOpts, arg0)
}

// Validators is a free data retrieval call binding the contract method 0x35aa2e44.
//
// Solidity: function validators(uint256 ) view returns(address)
func (_Minter *MinterCaller) Validators(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "validators", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Validators is a free data retrieval call binding the contract method 0x35aa2e44.
//
// Solidity: function validators(uint256 ) view returns(address)
func (_Minter *MinterSession) Validators(arg0 *big.Int) (common.Address, error) {
	return _Minter.Contract.Validators(&_Minter.CallOpts, arg0)
}

// Validators is a free data retrieval call binding the contract method 0x35aa2e44.
//
// Solidity: function validators(uint256 ) view returns(address)
func (_Minter *MinterCallerSession) Validators(arg0 *big.Int) (common.Address, error) {
	return _Minter.Contract.Validators(&_Minter.CallOpts, arg0)
}

// VerifyMintingSigner is a free data retrieval call binding the contract method 0xdfac3db6.
//
// Solidity: function verifyMintingSigner(bytes32 tx_hash, uint256 block_no, uint256 amount, address to, uint8 v, bytes32 r, bytes32 s, address signer) view returns(bool)
func (_Minter *MinterCaller) VerifyMintingSigner(opts *bind.CallOpts, tx_hash [32]byte, block_no *big.Int, amount *big.Int, to common.Address, v uint8, r [32]byte, s [32]byte, signer common.Address) (bool, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "verifyMintingSigner", tx_hash, block_no, amount, to, v, r, s, signer)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyMintingSigner is a free data retrieval call binding the contract method 0xdfac3db6.
//
// Solidity: function verifyMintingSigner(bytes32 tx_hash, uint256 block_no, uint256 amount, address to, uint8 v, bytes32 r, bytes32 s, address signer) view returns(bool)
func (_Minter *MinterSession) VerifyMintingSigner(tx_hash [32]byte, block_no *big.Int, amount *big.Int, to common.Address, v uint8, r [32]byte, s [32]byte, signer common.Address) (bool, error) {
	return _Minter.Contract.VerifyMintingSigner(&_Minter.CallOpts, tx_hash, block_no, amount, to, v, r, s, signer)
}

// VerifyMintingSigner is a free data retrieval call binding the contract method 0xdfac3db6.
//
// Solidity: function verifyMintingSigner(bytes32 tx_hash, uint256 block_no, uint256 amount, address to, uint8 v, bytes32 r, bytes32 s, address signer) view returns(bool)
func (_Minter *MinterCallerSession) VerifyMintingSigner(tx_hash [32]byte, block_no *big.Int, amount *big.Int, to common.Address, v uint8, r [32]byte, s [32]byte, signer common.Address) (bool, error) {
	return _Minter.Contract.VerifyMintingSigner(&_Minter.CallOpts, tx_hash, block_no, amount, to, v, r, s, signer)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_Minter *MinterCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Minter.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_Minter *MinterSession) Version() (string, error) {
	return _Minter.Contract.Version(&_Minter.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_Minter *MinterCallerSession) Version() (string, error) {
	return _Minter.Contract.Version(&_Minter.CallOpts)
}

// Delegate is a paid mutator transaction binding the contract method 0x9fa6dd35.
//
// Solidity: function delegate(uint256 amount) returns(bool)
func (_Minter *MinterTransactor) Delegate(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Minter.contract.Transact(opts, "delegate", amount)
}

// Delegate is a paid mutator transaction binding the contract method 0x9fa6dd35.
//
// Solidity: function delegate(uint256 amount) returns(bool)
func (_Minter *MinterSession) Delegate(amount *big.Int) (*types.Transaction, error) {
	return _Minter.Contract.Delegate(&_Minter.TransactOpts, amount)
}

// Delegate is a paid mutator transaction binding the contract method 0x9fa6dd35.
//
// Solidity: function delegate(uint256 amount) returns(bool)
func (_Minter *MinterTransactorSession) Delegate(amount *big.Int) (*types.Transaction, error) {
	return _Minter.Contract.Delegate(&_Minter.TransactOpts, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) returns(bool)
func (_Minter *MinterTransactor) Deposit(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Minter.contract.Transact(opts, "deposit", amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) returns(bool)
func (_Minter *MinterSession) Deposit(amount *big.Int) (*types.Transaction, error) {
	return _Minter.Contract.Deposit(&_Minter.TransactOpts, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amount) returns(bool)
func (_Minter *MinterTransactorSession) Deposit(amount *big.Int) (*types.Transaction, error) {
	return _Minter.Contract.Deposit(&_Minter.TransactOpts, amount)
}

// Elect is a paid mutator transaction binding the contract method 0xfae34a46.
//
// Solidity: function elect(uint256 blockNo, address[] candidates) returns(bool)
func (_Minter *MinterTransactor) Elect(opts *bind.TransactOpts, blockNo *big.Int, candidates []common.Address) (*types.Transaction, error) {
	return _Minter.contract.Transact(opts, "elect", blockNo, candidates)
}

// Elect is a paid mutator transaction binding the contract method 0xfae34a46.
//
// Solidity: function elect(uint256 blockNo, address[] candidates) returns(bool)
func (_Minter *MinterSession) Elect(blockNo *big.Int, candidates []common.Address) (*types.Transaction, error) {
	return _Minter.Contract.Elect(&_Minter.TransactOpts, blockNo, candidates)
}

// Elect is a paid mutator transaction binding the contract method 0xfae34a46.
//
// Solidity: function elect(uint256 blockNo, address[] candidates) returns(bool)
func (_Minter *MinterTransactorSession) Elect(blockNo *big.Int, candidates []common.Address) (*types.Transaction, error) {
	return _Minter.Contract.Elect(&_Minter.TransactOpts, blockNo, candidates)
}

// Vote is a paid mutator transaction binding the contract method 0x220fb1f3.
//
// Solidity: function vote(bytes32 txHash, uint256 blockNo, uint256 amount, address to, uint8 v, bytes32 r, bytes32 s, address signer) returns(uint256)
func (_Minter *MinterTransactor) Vote(opts *bind.TransactOpts, txHash [32]byte, blockNo *big.Int, amount *big.Int, to common.Address, v uint8, r [32]byte, s [32]byte, signer common.Address) (*types.Transaction, error) {
	return _Minter.contract.Transact(opts, "vote", txHash, blockNo, amount, to, v, r, s, signer)
}

// Vote is a paid mutator transaction binding the contract method 0x220fb1f3.
//
// Solidity: function vote(bytes32 txHash, uint256 blockNo, uint256 amount, address to, uint8 v, bytes32 r, bytes32 s, address signer) returns(uint256)
func (_Minter *MinterSession) Vote(txHash [32]byte, blockNo *big.Int, amount *big.Int, to common.Address, v uint8, r [32]byte, s [32]byte, signer common.Address) (*types.Transaction, error) {
	return _Minter.Contract.Vote(&_Minter.TransactOpts, txHash, blockNo, amount, to, v, r, s, signer)
}

// Vote is a paid mutator transaction binding the contract method 0x220fb1f3.
//
// Solidity: function vote(bytes32 txHash, uint256 blockNo, uint256 amount, address to, uint8 v, bytes32 r, bytes32 s, address signer) returns(uint256)
func (_Minter *MinterTransactorSession) Vote(txHash [32]byte, blockNo *big.Int, amount *big.Int, to common.Address, v uint8, r [32]byte, s [32]byte, signer common.Address) (*types.Transaction, error) {
	return _Minter.Contract.Vote(&_Minter.TransactOpts, txHash, blockNo, amount, to, v, r, s, signer)
}

// Vote0 is a paid mutator transaction binding the contract method 0x359afa49.
//
// Solidity: function vote(bytes32 txHash, uint256 blockNo, uint256 amount, address to) returns(uint256)
func (_Minter *MinterTransactor) Vote0(opts *bind.TransactOpts, txHash [32]byte, blockNo *big.Int, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Minter.contract.Transact(opts, "vote0", txHash, blockNo, amount, to)
}

// Vote0 is a paid mutator transaction binding the contract method 0x359afa49.
//
// Solidity: function vote(bytes32 txHash, uint256 blockNo, uint256 amount, address to) returns(uint256)
func (_Minter *MinterSession) Vote0(txHash [32]byte, blockNo *big.Int, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Minter.Contract.Vote0(&_Minter.TransactOpts, txHash, blockNo, amount, to)
}

// Vote0 is a paid mutator transaction binding the contract method 0x359afa49.
//
// Solidity: function vote(bytes32 txHash, uint256 blockNo, uint256 amount, address to) returns(uint256)
func (_Minter *MinterTransactorSession) Vote0(txHash [32]byte, blockNo *big.Int, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Minter.Contract.Vote0(&_Minter.TransactOpts, txHash, blockNo, amount, to)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns(bool)
func (_Minter *MinterTransactor) Withdraw(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Minter.contract.Transact(opts, "withdraw", amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns(bool)
func (_Minter *MinterSession) Withdraw(amount *big.Int) (*types.Transaction, error) {
	return _Minter.Contract.Withdraw(&_Minter.TransactOpts, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns(bool)
func (_Minter *MinterTransactorSession) Withdraw(amount *big.Int) (*types.Transaction, error) {
	return _Minter.Contract.Withdraw(&_Minter.TransactOpts, amount)
}

