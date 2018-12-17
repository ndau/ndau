package main

import (
	"fmt"
	"os"

	"github.com/oneiro-ndev/ndau/pkg/txjson"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func main() {
	tmpl, err := generator.ParseTemplate()
	check(err)
	context, err := generator.MakeContext()
	check(err)
	check(generator.ApplyTemplate(tmpl, context))
}
