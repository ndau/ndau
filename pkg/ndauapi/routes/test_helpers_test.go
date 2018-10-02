package routes_test

import "flag"

var isIntegration bool

func init() {
	flag.BoolVar(&isIntegration, "integration", false, "opt into integration tests")
	flag.Parse()
}
