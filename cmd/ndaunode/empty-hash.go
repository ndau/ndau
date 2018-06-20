package main

import (
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"
)

// get the hash of an empty database
func getEmptyHash() string {
	// create an in-memory app
	// the ignored variable here is associated mocked data;
	// it's safe to ignore, because these mocks are immediately discarded
	config, _, err := config.MakeTmpMock("")
	check(err)
	app, err := ndau.NewApp("mem", *config)
	check(err)
	return app.HashStr()
}
