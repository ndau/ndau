package routes_test

import (
	"flag"
	"os"
)

var isIntegration bool
var ndauRPC string
var chaosRPC string

func init() {
	flag.BoolVar(&isIntegration, "integration", false, "opt into integration tests")
	flag.StringVar(&ndauRPC, "ndaurpc", "", "ndau rpc url")
	flag.StringVar(&chaosRPC, "chaosrpc", "", "chaos rpc url")
	flag.Parse()

	os.Setenv("NDAUAPI_NDAU_RPC_URL", ndauRPC)
	os.Setenv("NDAUAPI_CHAOS_RPC_URL", chaosRPC)
}
