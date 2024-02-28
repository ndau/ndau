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
	"math/big"
	"os"
	"regexp"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

type priv_validator_key struct {
	address string
	pub_key struct {
		keytype string
		value   string
	}
	priv_key struct {
		keytype string
		value   string
	}
}

// getTendermintPrivateKey - Get the Tendermint private validator key file contents
// for signing minter voting transactions. This is the same as the ndau node account
// ownership key, but that is no longer available (not persisted on disk). Reading
// this key file is the same thing Tendermint does, so this usage is no less secure.

func getTendermintPrivateKey() ([]byte, error) {

	var priv_key []byte             // The 64-byte ed25519 private key
	var pvk_json priv_validator_key // Contents of priv_validator_key.json

	// Get the Tendermint data directory and concatenate the key file. os.Getenv
	// doesn't return an error, just an empty string if the environment variable
	// is either missing or null. That's strange (for an ndau node), but not an error.

	tmDataDir := os.Getenv("TM_DATA_DIR")
	pkFileName := tmDataDir + "/config/priv_validator_key.json"

	// Read the JSON key file and extract the private validator key

	pkJSON, err := os.ReadFile(pkFileName)

	if err == nil {

		// Unmarshal the JSON into the pk file struct and decode the base64
		// private key. We don't care about anything else since we're using it
		// for an Ethereum secp256k1 keypair. It's just 64 random bytes.

		err = json.Unmarshal(pkJSON, &pvk_json)
		priv_key, err = b64.StdEncoding.DecodeString(pvk_json.priv_key.value)
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

func makeSecp256k1Key(seed []byte) (*ecdsa.PrivateKey, error) {

	c := crypto.S256()

	seed, err := getTendermintPrivateKey()
	if err != nil {
		return nil, err
	}

	ethPrivKey, err := ECDSALegacy(c, seed)
	if err != nil {
		return nil, err
	}

	return ethPrivKey, nil
}

// isValid - check whether a string is a valid Ethereum address and is
// not a smart contract address

func isValid(ethaddr string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	valid := re.MatchString(ethaddr)

	// If it's a valid address, check to see if there's any bytecode there.
	// If not, it's a valid address for minting.

	if valid {
		address := common.HexToAddress(ethaddr)

		// TODO - Deal with RPC provider configuration

		client, err := ethclient.Dial("https://mainnet.infura.io")
		if err != nil {
			log.Fatal(err)
		}

		bytecode, err := client.CodeAt(context.Background(), address, nil) // nil is latest block
		if err != nil {
			log.Fatal(err)
		}

		if len(bytecode) == 0 {
			return true
		}
	}
	return false
}

func signAndSend(message string) error {
	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	// TODO - Deal with RPC provider configuration

	client, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		log.Fatal(err)
	}

	chainID, _ := client.ChainID(context.Background())

	// TODO - Generate fromAddress using Tendermint keypair

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	toAddress := common.HexToAddress("0x147B8eb97fD247D06C4006D269c90C1908Fb5D54")

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	// TODO - Put real values here

	var data []byte
	value := big.NewInt(0) // in wei

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:     fromAddress,
		To:       &toAddress,
		Gas:      uint64(0),
		GasPrice: gasPrice,
		Data:     data,
	})
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func submitMinterVote(ndau uint64, ethaddr string) error {
	if !isValid(ethaddr) {
		log.Fatal(0)
	}

	signAndSend("TODO - minter message here")

	return nil
}