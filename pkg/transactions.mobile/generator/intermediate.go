package generator

import (
	"fmt"
	"strings"
)

func nothing(s string) string { return s }

// Field stores metadata about a field
type Field struct {
	Name                     string
	Type                     string
	MobileType               string
	ConvertToNative          string
	FallibleNativeConversion bool
	ConvertToMobile          string
	ConstructorExcluded      bool
	AssignmentErrHandler     string
}

// NewField creates a new Field struct
func NewField(name, nativetype, mobiletype string) Field {
	return Field{
		Name:       name,
		Type:       nativetype,
		MobileType: mobiletype,
	}
}

// ExcludeFromConstructor excludes this field from constructors
//
// This is most useful for the list of signatures
func (f Field) ExcludeFromConstructor() Field {
	f.ConstructorExcluded = true
	return f
}

// ConvertNativeSimple adds native conversion for simple types
//
// This is appropriate for types like Ndau, which can simply be typecast from
// mobile representations such as int.
//
// `conversion` must be a go string with a single "%s" in it. This will be
// replaced at template time with the appropriate identifier.
//
// For example, for the Qty field of a Transfer, an appropriate value for
// `conversion` would be `"math.Ndau(%s)"`
func (f Field) ConvertNativeSimple(conversion string) Field {
	f.ConvertToNative = fmt.Sprintf(conversion, strings.ToLower(f.Name))
	return f
}

// ConvertNativeComplex adds native conversion for complex types
//
// This is appropriate for types like *keyaddr.Address, which must be passed
// through a process which might generate an error.
//
// `conversion` must be a go string with a single "%s" in it. This will be
// replaced at template time with the appropriate identifier.
//
// For example, for the Source field of a Transfer, an appropriate value for
// `conversion` would be `"address.Validate(%s.Address)"`
//
// Error handling is inserted automatically.
func (f Field) ConvertNativeComplex(conversion string) Field {
	f.ConvertToNative = fmt.Sprintf(conversion, strings.ToLower(f.Name))
	f.AssignmentErrHandler = fmt.Sprintf("if err != nil { return nil, errors.Wrap(err, \"%s\") }\n", strings.ToLower(f.Name))
	f.FallibleNativeConversion = true
	return f
}

// ConvertMobile adds mobile conversion
//
// `conversion` must be a go string with a single "%s" in it. This will be
// replaced at template time with the appropriate identifier.
//
// For example, for the Source field of a Transfer, an appropriate value for
// `conversion` would be `"keyaddr.Address{Address: %s.String()}"`
func (f Field) ConvertMobile(conversion string) Field {
	f.ConvertToMobile = fmt.Sprintf(conversion, "tx.tx."+f.Name)
	return f
}

// IsSlice is true if this field is a slice
func (f Field) IsSlice() bool {
	return strings.HasPrefix(f.Type, "[]")
}

// NameNative returns a unique identifier for the native version of a field
func (f Field) NameNative() string {
	return fmt.Sprintf("%sN", strings.ToLower(f.Name))
}

// AssignmentStmt produces the correct assignment operator
func (f Field) AssignmentStmt() string {
	if f.FallibleNativeConversion {
		return ", err := "
	}
	return " := "
}

// Transaction stores metadata about a transaction
type Transaction struct {
	Name    string
	Comment string
	Fields  []Field
}
