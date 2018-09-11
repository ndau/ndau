package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
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

type value struct {
	V        json.RawMessage `json:"v"`
	Encoding string          `json:"type"`
}

func (v value) json() []byte {
	js, err := json.Marshal(v)
	check(err)
	return js
}

func tryDecode(slice []byte) value {
	// try to interpret the first as msgp, then a raw string, then bytes
	buffer := new(bytes.Buffer)
	_, err := msgp.UnmarshalAsJSON(buffer, slice)
	if err == nil {
		return value{V: json.RawMessage(buffer.Bytes()), Encoding: "msgp"}
	}
	if utf8.ValidString(string(slice)) {
		return value{V: json.RawMessage(slice), Encoding: "utf8"}
	}
	return value{V: json.RawMessage(fmt.Sprintf("\"%s\"", base64.StdEncoding.EncodeToString(slice))), Encoding: "raw"}
}

func intoJSON(mock config.ChaosMock) string {
	buffer := new(bytes.Buffer)
	buffer.WriteRune('[')
	nsIdx := 0
	for outer, inner := range mock {
		if nsIdx != 0 {
			buffer.WriteRune(',')
		}
		nsIdx++
		buffer.WriteRune('{')
		buffer.WriteString("\"namespace\":")
		buffer.Write(tryDecode([]byte(outer)).json())
		buffer.WriteString(",\"data\":[")
		innerIdx := 0
		for key, value := range inner {
			if innerIdx != 0 {
				buffer.WriteRune(',')
			}
			innerIdx++
			buffer.WriteString("{\"key\":")
			buffer.Write(tryDecode([]byte(key)).json())
			buffer.WriteString(",\"value\":")
			buffer.Write(tryDecode(value).json())
			buffer.WriteRune('}')
		}
		buffer.WriteString("]}")
	}
	buffer.WriteRune(']')
	return string(buffer.Bytes())
}

func main() {
	conf, err := config.LoadDefault(config.DefaultConfigPath(getNdauhome()))
	check(err)

	if conf.UseMock == nil || *conf.UseMock == "" {
		check(errors.New("conf.UseMock not set"))
	}

	mock, err := config.LoadMock(*conf.UseMock)
	check(err)
	fmt.Println(intoJSON(mock))
}
