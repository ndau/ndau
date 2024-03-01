package ndau

import (
	"testing"
)

// TestIsValid - Check address validation

func TestIsValid(t *testing.T) {
	validAddr := "0x12ae66cdc592e10b60f9097a7b0d3c59fce29876"
	validAddrChecksum := "0x12AE66CDc592e10B60f9097a7b0D3C59fce29876"
	invalidContractAddress := "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"

	if !isValid(validAddr) {
		t.Fatalf(`valid address %s reported to be not valid`, validAddr)
	}

	if !isValid(validAddrChecksum) {
		t.Fatalf(`valid checksummed address %s reported to be not valid`, validAddrChecksum)
	}

	if isValid(invalidContractAddress) {
		t.Fatalf(`invalid smart contract address %s reported to be valid`, invalidContractAddress)
	}
}

func TestIsNotValid(t *testing.T) {
	invalidAddr := "0x12ae66cdc592e10b60f9097a7b0d3c59fce298762"
	invalidAddrChecksum := "0x22AE66CDc592e10B60f9097a7b0D3C59fce29876"
	invalidContractAddress := "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"

	if isValid(invalidAddr) {
		t.Fatalf(`invalid address %s reported to be valid`, invalidAddr)
	}

	if isValid(invalidAddrChecksum) {
		t.Fatalf(`invalid checksummed address %s reported to be valid`, invalidAddrChecksum)
	}

	if isValid(invalidContractAddress) {
		t.Fatalf(`invalid smart contract address %s reported to be valid`, invalidContractAddress)
	}
}

func TestTendermintPK(t *testing.T) {
	pk, err := getTendermintPrivateKey()
	if pk == nil || err != nil {
		t.Fatalf(`private validator key could not be read, %s`, err)
	}
}

func TestECDSAKeyGeneration(t *testing.T) {
	pk, err := getTendermintPrivateKey()
	if pk == nil || err != nil {
		t.Fatalf(`private validator key could not be read, %s`, err)
	}
	ECDSAKey, err := makeSecp256k1Key(pk)
	if ECDSAKey == nil || err != nil {
		t.Fatalf(`could not generate ECDSA key %s`, err)
	}
}
