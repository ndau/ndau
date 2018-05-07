package main

import (
	"fmt"
	"os"

	"gitlab.ndau.tech/experiments/ndau-chain/pkg/ndau"
)

// get the hash of an empty database
func getEmptyHash() string {
	// create an in-memory app
	app, err := ndau.NewApp("mem")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return app.HashStr()
}
