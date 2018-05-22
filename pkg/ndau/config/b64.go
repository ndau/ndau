package config

import (
	"bytes"
	"encoding/base64"
)

// NewB64Data creates a new B64data object
func NewB64Data(data []byte) B64Data {
	return B64Data(data)
}

// Bytes returns the bytes of the B64Data
func (b *B64Data) Bytes() []byte {
	return []byte(*b)
}

// UnmarshalText satisfies the encoding.TextUnmarshaler interface
func (b *B64Data) UnmarshalText(text []byte) error {
	bytes, err := base64.StdEncoding.DecodeString(string(text))
	if err == nil {
		*b = bytes
	}
	return err
}

// MarshalText satisfies the encoding.TextMarshaler interface
func (b B64Data) MarshalText() (text []byte, err error) {
	text = []byte(base64.StdEncoding.EncodeToString(b.Bytes()))
	return
}

// Equal returns true when `b` and `other` are equal
func (b *B64Data) Equal(other *B64Data) bool {
	if b == nil && other == nil {
		return true
	}
	if b == nil || other == nil {
		return false
	}
	return bytes.Equal(b.Bytes(), other.Bytes())
}
