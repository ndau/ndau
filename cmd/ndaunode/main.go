package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oneiro-ndev/ndau-chain/pkg/ndau"
	"github.com/oneiro-ndev/ndau-chain/pkg/ndau/config"
	"github.com/tendermint/abci/server"
	"github.com/tendermint/tmlibs/log"
)

var makeMocks = flag.Bool("make-mocks", false, "if set, make mock config data and exit")
var useNh = flag.Bool("use-ndauhome", false, "if set, keep database within $NDAUHOME/ndau")
var dbspec = flag.String("spec", "", "manually set the noms db spec")
var socketAddr = flag.String("addr", "0.0.0.0:46658", "socket address for incoming connection from tendermint")
var echoSpec = flag.Bool("echo-spec", false, "if set, echo the DB spec used and then quit")
var echoEmptyHash = flag.Bool("echo-empty-hash", false, "if set, echo the hash of the empty DB and then quit")

func getNdauhome() string {
	nh := os.ExpandEnv("$NDAUHOME")
	if len(nh) > 0 {
		return nh
	}
	return filepath.Join(os.ExpandEnv("$HOME"), ".ndau")
}

func getNdauConfigDir() string {
	return filepath.Join(getNdauhome(), "ndau")
}

func getDbSpec() string {
	if len(*dbspec) > 0 {
		return *dbspec
	}
	if *useNh {
		return filepath.Join(getNdauConfigDir(), "data")
	}
	// default to noms server for dockerization
	return "http://noms:8000"
}

func main() {
	flag.Parse()

	if *echoSpec {
		fmt.Println(getDbSpec())
		os.Exit(0)
	}

	if *echoEmptyHash {
		fmt.Println(getEmptyHash())
		os.Exit(0)
	}

	if *makeMocks {
		ndauhome := getNdauhome()
		configPath := config.DefaultConfigPath(ndauhome)
		mockPath := config.DefaultMockPath(ndauhome)
		fmt.Printf("Config: %s\n", configPath)
		fmt.Printf("Mock:   %s\n", mockPath)
		err := config.MakeMock(configPath, mockPath)
		if err == nil {
			os.Exit(0)
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("bin", "ndaunode")

	// JSG get socket addr from flag or default: 0.0.0.0:46658
	sa := *socketAddr

	app, err := ndau.NewApp(getDbSpec())
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	app.SetLogger(logger)
	app.LogState()

	server := server.NewSocketServer(sa, app)
	server.SetLogger(logger)

	err = server.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	logger.Info("started ABCI socket server", "address", sa, "name", server.String())
	// we want to keep this service running indefinitely
	// if there were more commands to run, we'd probably want to split this into a separate
	// goroutine and deal with closing options, but for now, it's probably fine to actually
	// just let the main routine hang here
	<-server.Quit()
}
