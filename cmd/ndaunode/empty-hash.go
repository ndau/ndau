package main

import (
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"
)

// get the hash of an empty database
func getEmptyHash() string {
	// create an in-memory app
	config, err := config.MakeTmpMock("")
	check(err)
	app, err := ndau.NewApp("mem", *config)
	check(err)
	return app.HashStr()
}
