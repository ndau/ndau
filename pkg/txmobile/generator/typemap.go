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
		f.MobileType = "string"
		f.ConvertToMobile = func(s string) string {
			return fmt.Sprintf(
				"%s.String()",
				s,
			)
		}
		f.ConvertToNative = func(s string) string {
			return fmt.Sprintf(
				"address.Validate(%s)",
				s,
			)
		}
		f.FallibleNativeConversion = true

	case "signature.PublicKey", "[]signature.PublicKey":
		f.MobileType = "string"
		f.ConvertToMobile = func(s string) string {
			return fmt.Sprintf(
				"%s.MarshalString()",
				s,
			)
		}
		f.FallibleMobileConversion = true

		f.ConvertToNative = func(s string) string {
			return fmt.Sprintf("signature.ParsePublicKey(%s)", s)
		}
		f.FallibleNativeConversion = true
		f.PointerNativeConversion = true

	case "signature.Signature", "[]signature.Signature":
		f.MobileType = "string"
		f.ConvertToMobile = func(s string) string {
			return fmt.Sprintf(
				"%s.MarshalString()",
				s,
			)
		}
		f.FallibleMobileConversion = true

		f.ConvertToNative = func(s string) string {
			return fmt.Sprintf("signature.ParseSignature(%s)", s)
		}
		f.FallibleNativeConversion = true
		f.PointerNativeConversion = true

	default:
		return fmt.Errorf("unknown type: %s", f.Type)
	}

	if f.Name == "Signatures" || f.Name == "Signature" {
		f.ConstructorExcluded = true
		// signatures get special setters, so we don't want the generic kind
	} else if f.Type == "[]signature.PublicKey" || f.Type == "[]signature.Signature" {
		// can't just do `if strings.HasPrefix(f.Type, "[]")``, because that would
		// erroneously hit `[]byte`
		//
		// this must be an "else if" instead of an isolated "if" or "switch" so that
		// we don't end up with redundant "appendsignature" methods
		f.ConstructorExcluded = true
		f.MakeSetter = true
	}

	return nil
}
