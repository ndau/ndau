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
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	minter "github.com/ndau/ndau/pkg/ndau/minter"
)

// Tendermint priv_validator_key.json structure
type PVKey struct {
	Address string `json:"address"`
	Pub_key struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"pub_key"`
	Priv_key struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"priv_key"`
}

// getTendermintPrivateKey - Get the Tendermint private validator key file contents
// for signing minter voting transactions. This is the same as the ndau node account
// ownership key, but that is no longer available (not persisted on disk). Reading
// this key file is the same thing Tendermint does, so this usage is no less secure.

func GetTendermintPrivateKey() ([]byte, error) {

	var priv_key []byte // The 64-byte ed25519 private key

	tmDataDir := os.Getenv("TM_DATA_DIR")

	if tmDataDir == "" {
		tmDataDir = "/Users/edmcnierney/go/src/github.com/ndau/edmcnierney/node_identities/abundance/tendermint"
	}

	// Read the JSON key file and extract the private validator key

	pvkJSON, err := os.ReadFile(tmDataDir + "/config/priv_validator_key.json")
	if err != nil {
		return nil, err
	}

	pvk := PVKey{}

	// Unmarshal the JSON into the pk file struct and decode the base64
	// private key. We don't care about anything else since we're using it
	// for an Ethereum secp256k1 keypair. It's just 64 random bytes.

	err = json.Unmarshal([]byte(pvkJSON), &pvk)
	if err == nil {
		priv_key, err = b64.StdEncoding.DecodeString(pvk.Priv_key.Value)
	}

	return priv_key, err
}

// ECDSALegacy - Create a deterministic secp256k1 key. This code is copied directly
// from https://github.com/FiloSottile/keygen, the only change being the use of a
// byte array as the seed rather than wrapping an io.Reader interface around it.

func ECDSALegacy(c elliptic.Curve, b []byte) (*ecdsa.PrivateKey, error) {

	params := c.Params()

	seedlen := params.N.BitLen()/8 + 8
	one := big.NewInt(1)
	k := new(big.Int).SetBytes(b[0 : seedlen-1])
	n := new(big.Int).Sub(params.N, one)
	k.Mod(k, n)
	k.Add(k, one)

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = c
	priv.D = k
	priv.PublicKey.X, priv.PublicKey.Y = c.ScalarBaseMult(k.Bytes())
	return priv, nil
}

// makeSecp256k1Key - Use the Tendermint private key bytes to generate a crypto.ecdsa key

func MakeSecp256k1Key(seed []byte) (*ecdsa.PrivateKey, error) {

	c := crypto.S256()

	ethPrivKey, err := ECDSALegacy(c, seed)
	if err != nil {
		return nil, err
	}

	return ethPrivKey, nil
}

// isValid - check whether a string is a valid Ethereum address and is
// not a smart contract address. Don't do any fancy error handling - if anything
// goes wrong return false. If it's an RPC provider glitch the user can just
// try again later; we probably can't submit the minter vote, either.

func IsEthAddressValid(ethaddr string) error {

	// Basic format check

	errorNotValid := errors.New("address is not valid")

	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	if !re.MatchString(ethaddr) {
		return errorNotValid
	}

	// We know ethaddr is a valid hex string of the right length. If it
	// looks like it's checksummed (has uppercase characters), test it.

	address := common.HexToAddress(ethaddr)

	if strings.ToLower(ethaddr) != ethaddr {
		mcaddress, err := common.NewMixedcaseAddressFromString(ethaddr)
		if err != nil || !mcaddress.ValidChecksum() {
			return errorNotValid
		}
	}

	// If it's a valid address, check to see if there's any bytecode there.
	// If not, it's a valid address for minting.

	// TODO - Deal with RPC provider configuration

	rpc := os.Getenv("RPC_PROVIDER")
	if rpc == "" {
		rpc = "https://mainnet.infura.io/v3/2d964329cb8746139ba47fe1ccf3b9e5"
	}

	client, err := ethclient.Dial(rpc)
	if err != nil {
		return errorNotValid
	}

	bytecode, err := client.CodeAt(context.Background(), address, nil) // nil is latest block
	if err != nil {
		return errorNotValid
	}

	if len(bytecode) != 0 { // There's a smart contract there
		return errorNotValid
	}

	return nil
}

func MintNPAY(hash [32]byte, blockNo *big.Int, amount *big.Int, ethAddr string) error {
	if IsEthAddressValid(ethAddr) != nil {
		return errors.New("ethereum address is not valid")
	}

	tmPk, err := GetTendermintPrivateKey()
	if err != nil {
		return errors.New("could not retrieve Tendermint private validator key")
	}

	pk, err := MakeSecp256k1Key(tmPk)
	if err != nil {
		return errors.New("couldn't create secp256k1 key")
	}

	publicKey := pk.Public()

	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("public key is not *ecdsa.PublicKey")
	}

	ethereumRPC := os.Getenv("ETH_RPC_ENDPOINT")
	if ethereumRPC == "" {
		ethereumRPC = "https://goerli.infura.io/v3/2d964329cb8746139ba47fe1ccf3b9e5"
	}

	client, err := ethclient.Dial(ethereumRPC)
	if err != nil {
		return err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	toAddress := common.HexToAddress(ethAddr)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	auth, _ := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(1))
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	// The below is correct Minter address on Goerli already
	address := common.HexToAddress("0x38344b0F40D7e558d6DF24C7f04Fd57CA5979d60")
	instance, err := minter.NewMinter(address, client)
	if err != nil {
		log.Fatal(err)
	}

	// txHash, _ := uint256.FromHex("3f8c971c7082894193982104b01b6c9bc786fa70515460fa2de2b36ef4866253")
	// amount :=new(big.Int)
	// amount.SetString("de0b6b3a7640000",16) // 1 NPAY

	tx, err := instance.Vote0(auth, hash, blockNo, amount, toAddress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(tx)

	return nil
}
