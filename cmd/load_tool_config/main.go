package main

import (
	"fmt"
	"os"

	tc "github.com/oneiro-ndev/ndaunode/pkg/tool.config"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func main() {
	config, err := tc.LoadDefault(tc.GetConfigPath())
	check(err)
	fmt.Println("successfully loaded config from:")
	fmt.Println(tc.GetConfigPath())
	err = config.Save()
	check(err)
	fmt.Println("successfully saved config")
}
