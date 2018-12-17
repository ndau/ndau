package generator

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// Field stores metadata about a field
type Field struct {
	Name            string
	Type            string
	Literal         string
	FallibleLiteral bool
	PointerLiteral  bool
	Transaction     *Transaction
}

// NewField creates a new Field struct
func NewField(name, nativeType string, tx *Transaction) Field {
	f := Field{
		Name:        name,
		Type:        nativeType,
		Transaction: tx,
	}
	f.fillFieldFromType()
	return f
}

// We can get the list of types in transactions.go in bash:
//
//     bat pkg/ndau/transactions.go | rg -vF '//' | rg -v '^\}?$' | rg -vF type | rg -vF 'var _' | tr -s '\t' ' ' | cut -d' ' -f3 | sort -u
//
// which produces some chaff, and then the types implemented below:

func randBytes(qty int) ([]byte, error) {
	buffer := make([]byte, qty)
	_, err := rand.Read(buffer)
	return buffer, err
}

func byteS(b byte) string {
	return fmt.Sprintf("0x%02x", b)
}

func bytesS(bs []byte) string {
	bS := make([]string, len(bs))
	for idx := range bs {
		bS[idx] = byteS(bs[idx])
	}
	return fmt.Sprintf("[]byte{%s}", strings.Join(bS, ", "))
}

func (f *Field) nlit(v int64) {
	f.Literal = fmt.Sprintf("%s(%d)", f.Type, v)
}

func (f *Field) fillFieldFromType() error {
	switch f.Type {
	case "string":
		charset := []byte("abcdefghijklmnopqrstuvwxyz")
		words := 1 + rand.Intn(10) // rand value in [1..10]
		out := []byte("string: ")

		for word := 0; word < words; word++ {
			chars := 1 + rand.Intn(8) // rand value in [1..8]
			for char := 0; char < chars; char++ {
				out = append(out, charset[rand.Intn(len(charset))])
			}
			out = append(out, ' ')
		}

		f.Literal = fmt.Sprintf("\"%s\"", out)

	case "int64":
		f.nlit(rand.Int63())
	case "uint64":
		f.nlit(int64(rand.Uint32()))
	case "math.Ndau":
		f.nlit(rand.Int63n(constants.NapuPerNdau * 100000))
	case "math.Duration":
		f.nlit(rand.Int63n(math.Year * 5))

	case "byte":
		bytes, err := randBytes(1)
		if err != nil {
			return err
		}
		f.Literal = byteS(bytes[0])

	case "[]byte":
		bytes, err := randBytes(address.MinDataLength)
		if err != nil {
			return err
		}

		f.Literal = bytesS(bytes)

	case "address.Address":
		f.FallibleLiteral = true
		bytes, err := randBytes(address.MinDataLength)
		if err != nil {
			return err
		}
		addr, err := address.Generate(address.KindUser, bytes)
		if err != nil {
			return err
		}
		f.Literal = fmt.Sprintf("address.Validate(\"%s\")", addr)

	case "signature.PublicKey", "[]signature.PublicKey":
		f.FallibleLiteral = true
		f.PointerLiteral = true
		bytes, err := randBytes(signature.Ed25519.PublicKeySize())
		if err != nil {
			return err
		}
		f.Literal = fmt.Sprintf("signature.RawPublicKey(signature.Ed25519, %s, nil)", bytesS(bytes))

	case "signature.Signature", "[]signature.Signature":
		f.FallibleLiteral = true
		f.PointerLiteral = true
		bytes, err := randBytes(signature.Ed25519.SignatureSize())
		if err != nil {
			return err
		}
		f.Literal = fmt.Sprintf("signature.RawSignature(signature.Ed25519, %s)", bytesS(bytes))

	default:
		return fmt.Errorf("unknown type: %s", f.Type)
	}

	return nil
}

// LiteralName returns an appropriate literal name for an instance of this field
func (f *Field) LiteralName() string {
	return fmt.Sprintf(
		"%s%s",
		strings.ToLower(f.Transaction.Name),
		strings.Title(f.Name),
	)
}

// IsSlice is true when the field type is a slice
func (f *Field) IsSlice() bool {
	return strings.HasPrefix(f.Type, "[]")
}
