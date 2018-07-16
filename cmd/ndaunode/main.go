package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"
	"github.com/tendermint/tendermint/abci/server"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

var makeMocks = flag.Bool("make-mocks", false, "if set, make mock config data and exit")
var makeChaosMocks = flag.Bool("make-chaos-mocks", false, "if set, make mock data on the chaos chain and exit")
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

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
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

	ndauhome := getNdauhome()
	configPath := config.DefaultConfigPath(ndauhome)
	if *makeMocks {
		generateMocks(ndauhome, configPath)
	}

	if *makeChaosMocks {
		generateChaosMocks(ndauhome, configPath)
	}

	conf, err := config.LoadDefault(configPath)
	check(err)

	app, err := ndau.NewApp(getDbSpec(), *conf)
	check(err)

	logger := app.GetLogger()
	logger = logger.WithField("bin", "chaosnode")
	app.SetLogger(logger)
	app.LogState()

	server := server.NewSocketServer(*socketAddr, app)
	server.SetLogger(tmlog.NewTMLogger(os.Stderr))

	err = server.Start()
	check(err)

	logger.Info(
		"started ABCI socket server",
		"address", *socketAddr,
		"name", server.String(),
	)
	// we want to keep this service running indefinitely
	// if there were more commands to run, we'd probably want to split this into a separate
	// goroutine and deal with closing options, but for now, it's probably fine to actually
	// just let the main routine hang here
	<-server.Quit()
}
