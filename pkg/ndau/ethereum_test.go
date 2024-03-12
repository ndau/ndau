package ndau

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

// TestIsEthAddressValid - Check address validation

func TestIsEthAddressValid(t *testing.T) {
	validAddr := "0x12ae66cdc592e10b60f9097a7b0d3c59fce29876"
	validAddrChecksum := "0x12AE66CDc592e10B60f9097a7b0D3C59fce29876"
	invalidContractAddress := "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"

	if IsEthAddressValid(validAddr) != nil {
		t.Fatalf(`valid address %s reported to be not valid`, validAddr)
	}

	if IsEthAddressValid(validAddrChecksum) != nil {
		t.Fatalf(`valid checksummed address %s reported to be not valid`, validAddrChecksum)
	}

	if IsEthAddressValid(invalidContractAddress) == nil {
		t.Fatalf(`invalid smart contract address %s reported to be valid`, invalidContractAddress)
	}
}

func TestIsNotValid(t *testing.T) {
	invalidAddr := "0x12ae66cdc592e10b60f9097a7b0d3c59fce298762"
	invalidAddrChecksum := "0x22AE66CDc592e10B60f9097a7b0D3C59fce29876"
	invalidContractAddress := "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"

	if IsEthAddressValid(invalidAddr) == nil {
		t.Fatalf(`invalid address %s reported to be valid`, invalidAddr)
	}

	if IsEthAddressValid(invalidAddrChecksum) == nil {
		t.Fatalf(`invalid checksummed address %s reported to be valid`, invalidAddrChecksum)
	}

	if IsEthAddressValid(invalidContractAddress) == nil {
		t.Fatalf(`invalid smart contract address %s reported to be valid`, invalidContractAddress)
	}
}

func TestTendermintPK(t *testing.T) {
	pk, err := GetTendermintPrivateKey()
	if pk == nil || err != nil {
		t.Fatalf(`private validator key could not be read, %s`, err)
	}
}

func TestECDSAKeyGeneration(t *testing.T) {
	pk, err := GetTendermintPrivateKey()
	if pk == nil || err != nil {
		t.Fatalf(`private validator key could not be read, %s`, err)
	}
	ECDSAKey, err := MakeSecp256k1Key(pk)
	if ECDSAKey == nil || err != nil {
		t.Fatalf(`could not generate ECDSA key %s`, err)
	}
}

func TestSignAndSend(t *testing.T) {
	var pk []byte
	var ECDSAKey *ecdsa.PrivateKey

	pk, err := GetTendermintPrivateKey()
	if err != nil {
		t.Fatalf(`could not get Tendermint private key`)
	}

	ECDSAKey, err = MakeSecp256k1Key(pk)
	if err != nil {
		t.Fatalf(`could not get private key`)
	}

	err = signAndSend([]byte("Vote for me!"), "0xb5300b33A656A291f4400D76fD9572011698EC71", ECDSAKey)
	if err != nil {
		t.Fatalf(`could not sign and send transaction: %s`, err)
	}
}

func TestMainnetValidators(t *testing.T) {
	var pk []byte
	var privateKey *ecdsa.PrivateKey

	pk, err := GetTendermintPrivateKey()
	if err != nil {
		t.Fatalf(`could not get Tendermint private key`)
	}

	privateKey, err = MakeSecp256k1Key(pk)
	if err != nil {
		t.Fatalf(`could not get private key`)
	}

	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKey)

	fmt.Println("Address is " + address.Hex())
	t.Log((address.Hex()))
}
