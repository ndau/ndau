package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
func getNdauhome() string {
	nh := os.ExpandEnv("$NDAUHOME")
	if len(nh) > 0 {
		return nh
	}
	return filepath.Join(os.ExpandEnv("$HOME"), ".ndau")
}

func tryDecode(slice []byte) string {
	// try to interpret the first as msgp, then a raw string, then bytes
	buffer := new(bytes.Buffer)
	_, err := msgp.UnmarshalAsJSON(buffer, slice)
	if err == nil {
		return fmt.Sprintf("%s (msgp)", string(buffer.Bytes()))
	}
	if utf8.ValidString(string(slice)) {
		return fmt.Sprintf("%s (utf8)", string(slice))
	}
	return fmt.Sprintf("%s (bytes)", base64.StdEncoding.EncodeToString(slice))
}

func main() {
	conf, err := config.LoadDefault(config.DefaultConfigPath(getNdauhome()))
	check(err)

	if conf.UseMock == "" {
		check(errors.New("conf.UseMock not set"))
	}

	mock, err := config.LoadMock(conf.UseMock)
	check(err)

	for outer, inner := range mock {
		fmt.Printf("Namespace: %s\n", tryDecode([]byte(outer)))
		for key, value := range inner {
			fmt.Println("  Key:   ", tryDecode([]byte(key)))
			fmt.Println("  Value: ", tryDecode(value))
		}
	}
}
