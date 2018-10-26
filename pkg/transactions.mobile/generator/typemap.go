package generator

import (
	"fmt"
	"strings"
)

// this file contains helpers used to map native to mobile-compatible types
//
// We can get the list of types in transactions.go in bash:
//
//     bat pkg/ndau/transactions.go | rg -vF '//' | rg -v '^\}?$' | rg -vF type | rg -vF 'var _' | tr -s '\t' ' ' | cut -d' ' -f3 | sort -u
//
// which produces some chaff, and then:
//
//     []byte
//     []signature.PublicKey
//     []signature.Signature
//     address.Address
//     int64
//     math.Duration
//     math.Ndau
//     signature.PublicKey
//     signature.Signature
//     string
//     uint64

func (f *Field) fillFieldFromType() error {
	flname := strings.ToLower(f.Name)
	switch f.Type {
	case "[]byte":
		f.MobileType = "string"
		f.ConvertToMobile = fmt.Sprintf(
			"base64.StdEncoding.EncodeToString(tx.tx.%s)",
			f.Name,
		)
		f.ConvertToNative = fmt.Sprintf(
			"base64.StdEncoding.DecodeString(%s)",
			flname,
		)
		f.FallibleNativeConversion = true

	case "[]signature.PublicKey":
		return fmt.Errorf("%s unimplemented", f.Type)
	case "[]signature.Signature":
		return fmt.Errorf("%s unimplemented", f.Type)
	case "address.Address":
		f.MobileType = "string"
		f.ConvertToMobile = fmt.Sprintf(
			"keyaddr.Address{Address: tx.tx.%s.String()}",
			f.Name,
		)
		f.ConvertToNative = fmt.Sprintf(
			"address.Validate(%s.Address)",
			flname,
		)
		f.FallibleNativeConversion = true
	case "int64", "uint64", "math.Ndau", "math.Duration":
		f.MobileType = "int64"
		f.ConvertToMobile = fmt.Sprintf("int64(tx.tx.%s)", f.Name)
		f.ConvertToNative = fmt.Sprintf("%s(%s)", f.Type, flname)
	case "signature.PublicKey":

	case "signature.Signature":

	case "string":

	default:
		return fmt.Errorf("unknown type: %s", f.Type)
	}
	return nil
}
