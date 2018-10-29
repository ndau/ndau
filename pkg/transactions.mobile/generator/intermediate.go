package generator

import (
	"fmt"
	"strings"
)

// Field stores metadata about a field
type Field struct {
	Name                     string
	Type                     string
	MobileType               string
	ConvertToNative          func(string) string
	FallibleNativeConversion bool
	PreCreateNative          bool
	ConvertToMobile          func(string) string
	FallibleMobileConversion bool
	ConstructorExcluded      bool
}

// NewField creates a new Field struct
func NewField(name, nativeType string) Field {
	f := Field{
		Name: name,
		Type: nativeType,
	}
	f.fillFieldFromType()
	if name == "Signatures" {
		f.ConstructorExcluded = true
	}
	return f
}

// AssignmentErrHandler is the err handler for fallible conversions as appropriate
func (f Field) AssignmentErrHandler() string {
	if f.FallibleNativeConversion {
		return fmt.Sprintf("if err != nil { return nil, errors.Wrap(err, \"%s\") }\n", strings.ToLower(f.Name))
	}
	return ""
}

// NameNative returns an identifier for the native version of a field
func (f Field) NameNative() string {
	return fmt.Sprintf("%sN", strings.ToLower(f.Name))
}

// NameSlice returns an identifier for a slice version of a field
func (f Field) NameSlice() string {
	return fmt.Sprintf("%sS", strings.ToLower(f.Name))
}

// AssignmentStmt produces the correct assignment operator
func (f Field) AssignmentStmt(fallible, precreate bool) string {
	operator := ":="
	if precreate {
		operator = "="
	}

	if fallible {
		return fmt.Sprintf(", err %s ", operator)
	}
	return fmt.Sprintf(" %s ", operator)
}

// Transaction stores metadata about a transaction
type Transaction struct {
	Name    string
	Comment string
	Fields  []Field
}
