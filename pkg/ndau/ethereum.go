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
	"math/big"
	"regexp"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

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
