package mobile

// generated with github.com/oneiro-ndev/ndau/pkg/transactions.mobile/generator
// DO NOT EDIT

import (
	"encoding/base64"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/keyaddr"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau CommandValidatorChange transaction
// for use in React and in particular react-native.
//
// It is meant to be built using the gomobile tool, so the API is constrained
// to particular types of parameters:
//
// * string
// * signed integer and floating point types
// * []byte
// * functions with specific restrictions
// * structs and interfaces consisting of only these types
//
// Unfortunately, react-native puts additional requirements that makes `[]byte`
// particularly challenging to use. To the degree possible, we take advantage
// of types' `(Un)MarshalText` implementations to generate and parse strings.
// Where that's impossible, we use the standard base64 encoding of the binary
// representation of the type.
//
// This package, therefore, consists mainly of wrappers so that we don't have
// to modify our idiomatic Go code to conform to these requirements.

// CommandValidatorChange is a mobile compatible wrapper for a CommandValidatorChange transaction
type CommandValidatorChange struct {
	tx ndau.CommandValidatorChange
}

// NewCommandValidatorChange constructs a new unsigned CommandValidatorChange transaction
func NewCommandValidatorChange(
	publickey string,
	power int64,
	sequence int64,
) (*CommandValidatorChange, error) {
	publickeyN, err := base64.StdEncoding.DecodeString(publickey)
	if err != nil {
		return nil, errors.Wrap(err, "publickey")
	}

	return &CommandValidatorChange{
		tx: ndau.CommandValidatorChange{
			PublicKey: publickeyN,
			Power:     int64(power),
			Sequence:  uint64(sequence),
		},
	}, nil
}

// ParseCommandValidatorChange parses a string into a CommandValidatorChange, if possible
func ParseCommandValidatorChange(s string) (*CommandValidatorChange, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseCommandValidatorChange: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseCommandValidatorChange: unmarshal")
	}
	trp, isTr := tx.(*ndau.CommandValidatorChange)
	if !isTr {
		return nil, errors.New("ParseCommandValidatorChange: transactable was not CommandValidatorChange")
	}

	return &CommandValidatorChange{tx: *trp}, nil
}

// ToString produces the b64 encoding of the bytes of the transaction
func (tx *CommandValidatorChange) ToString() (string, error) {
	if tx == nil {
		return "", errors.New("nil commandvalidatorchange")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "commandvalidatorchange: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetPublicKey gets the publickey of the CommandValidatorChange
//
// Returns `nil` if CommandValidatorChange is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *CommandValidatorChange) GetPublicKey() *string {
	if tx == nil {
		return nil
	}
	publickey := base64.StdEncoding.EncodeToString(tx.tx.PublicKey)

	return &publickey
}

// GetPower gets the power of the CommandValidatorChange
//
// Returns `nil` if CommandValidatorChange is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *CommandValidatorChange) GetPower() *int64 {
	if tx == nil {
		return nil
	}
	power := int64(tx.tx.Power)

	return &power
}

// GetSequence gets the sequence of the CommandValidatorChange
//
// Returns `nil` if CommandValidatorChange is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *CommandValidatorChange) GetSequence() *int64 {
	if tx == nil {
		return nil
	}
	sequence := int64(tx.tx.Sequence)

	return &sequence
}

// GetNumSignatures gets the number of signatures of the CommandValidatorChange
//
// If tx == nil, returns -1
func (tx *CommandValidatorChange) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this CommandValidatorChange
func (tx *CommandValidatorChange) GetSignature(idx int) (*keyaddr.Signature, error) {
	if tx == nil {
		return nil, errors.New("nil commandvalidatorchange")
	}
	if idx < 0 || idx >= len(tx.tx.Signatures) {
		return nil, errors.New("invalid index")
	}
	signature, err := keyaddr.SignatureFrom(tx.tx.Signatures[idx])
	if err != nil {
		return nil, errors.Wrap(err, "signatures")
	}

	return signature, nil
}

// SignableBytes returns the b64 encoding of the signable bytes of this commandvalidatorchange
func (tx *CommandValidatorChange) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil commandvalidatorchange")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this commandvalidatorchange
func (tx *CommandValidatorChange) AppendSignature(sig *keyaddr.Signature) error {
	if sig == nil {
		return errors.New("sig must not be nil")
	}
	sigS, err := sig.ToSignature()
	if err != nil {
		return errors.Wrap(err, "converting signature")
	}
	tx.tx.Signatures = append(tx.tx.Signatures, sigS)
	return nil
}

// Hash computes the hash of this commandvalidatorchange
func (tx *CommandValidatorChange) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *CommandValidatorChange) Name() string {
	if tx == nil {
		return ""
	}
	return "CommandValidatorChange"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *CommandValidatorChange) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
