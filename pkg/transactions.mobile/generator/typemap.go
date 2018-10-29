package generator

import (
	"fmt"
)

// this file contains helpers used to map native to mobile-compatible types
//
// We can get the list of types in transactions.go in bash:
//
//     bat pkg/ndau/transactions.go | rg -vF '//' | rg -v '^\}?$' | rg -vF type | rg -vF 'var _' | tr -s '\t' ' ' | cut -d' ' -f3 | sort -u
//
// which produces some chaff, and then the types implemented below:

func nothing(s string) string { return s }

func (f *Field) fillFieldFromType() error {
	f.ConvertToMobile = nothing
	f.ConvertToNative = nothing

	switch f.Type {
	case "string":
		f.MobileType = "string"

	case "int64", "uint64", "math.Ndau", "math.Duration":
		f.MobileType = "int64"
		f.ConvertToMobile = func(s string) string { return fmt.Sprintf("int64(%s)", s) }
		f.ConvertToNative = func(s string) string { return fmt.Sprintf("%s(%s)", f.Type, s) }

	case "[]byte":
		f.MobileType = "string"
		f.ConvertToMobile = func(s string) string {
			return fmt.Sprintf(
				"base64.StdEncoding.EncodeToString(%s)",
				s,
			)
		}
		f.ConvertToNative = func(s string) string {
			return fmt.Sprintf(
				"base64.StdEncoding.DecodeString(%s)",
				s,
			)
		}
		f.FallibleNativeConversion = true

	case "address.Address":
		f.MobileType = "*keyaddr.Address"
		f.ConvertToMobile = func(s string) string {
			return fmt.Sprintf(
				"keyaddr.Address{Address: %s.String()}",
				s,
			)
		}
		f.ConvertToNative = func(s string) string {
			return fmt.Sprintf(
				"address.Validate(%s.Address)",
				s,
			)
		}
		f.FallibleNativeConversion = true

	case "signature.PublicKey", "[]signature.PublicKey":
		f.MobileType = "*keyaddr.Key"
		f.ConvertToMobile = func(s string) string {
			return fmt.Sprintf(
				"keyaddr.KeyFromPublic(%s)",
				s,
			)
		}
		f.FallibleMobileConversion = true
		f.PointerMobileConversion = true

		f.ConvertToNative = func(s string) string { return fmt.Sprintf("%s.ToPublicKey()", s) }
		f.FallibleNativeConversion = true

	case "signature.Signature", "[]signature.Signature":
		f.MobileType = "*keyaddr.Signature"
		f.ConvertToMobile = func(s string) string {
			return fmt.Sprintf(
				"keyaddr.SignatureFrom(%s)",
				s,
			)
		}
		f.FallibleMobileConversion = true
		f.PointerMobileConversion = true

		f.ConvertToNative = func(s string) string { return fmt.Sprintf("%s.ToSignature()", s) }
		f.FallibleNativeConversion = true

	default:
		return fmt.Errorf("unknown type: %s", f.Type)
	}
	return nil
}
