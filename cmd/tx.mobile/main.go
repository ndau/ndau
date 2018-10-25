package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/oneiro-ndev/ndau/pkg/transactions.mobile/generator"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func main() {
	ast, err := generator.ParseTransactions()
	check(err)
	fmt.Println("parsed transactions")

	txIDs := generator.FindDefinition(ast, generator.TxIDs)
	if txIDs == nil {
		check(errors.New("TxIDs not found"))
	}

	fmt.Printf(
		"found TxIDs [%d:%d]: %s\n",
		txIDs.Definition.Pos(), txIDs.Definition.End(),
		txIDs.DefinitionType,
	)

	// emit the AST of TxIDs in a pretty-ish way
	// check(txIDs.Write(os.Stdout))

	txNames, err := generator.GetTxNames(txIDs.Definition)
	check(err)

	fmt.Printf("Found %d names:\n", len(txNames))
	for _, n := range txNames {
		fmt.Println("  ", n)
	}
}
