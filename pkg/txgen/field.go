package txgen

import (
	"fmt"
	"strings"
)

// Field stores metadata about a field
type Field struct {
	Name        string
	Type        string
	Transaction *Transaction
	Length      string
	Bytes       string
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

func (f *Field) fillFieldFromType() error {
	switch f.Type {
	case "string", "[]byte":
		f.Length = fmt.Sprintf("len(tx.%s)", f.Name)
		f.Bytes = fmt.Sprintf("[]byte(tx.%s)", f.Name)
	case "int64", "uint64", "math.Ndau", "math.Duration":
		f.Length = "8"
		f.Bytes = fmt.Sprintf("intbytes(uint64(tx.%s))", f.Name)
	case "byte":
		f.Length = "1"
		f.Bytes = fmt.Sprintf("[]byte{tx.%s}", f.Name)
	case "address.Address":
		f.Length = "address.AddrLength"
		f.Bytes = fmt.Sprintf("[]byte(tx.%s.String())", f.Name)
	case "signature.PublicKey", "[]signature.PublicKey", "signature.Signature", "[]signature.Signature":
		f.Length = fmt.Sprintf("tx.%s.MsgSize()", f.Name)
		f.Bytes = fmt.Sprintf("tx.%s.MarshalMsg(nil)", f.Name)

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
