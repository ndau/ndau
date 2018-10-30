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

// This file provides an interface to the Ndau NominateNodeReward transaction
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

// NominateNodeReward is a mobile compatible wrapper for a NominateNodeReward transaction
type NominateNodeReward struct {
	tx ndau.NominateNodeReward
}

// NewNominateNodeReward constructs a new unsigned NominateNodeReward transaction
func NewNominateNodeReward(
	random int64,
	sequence int64,
) (*NominateNodeReward, error) {
	return &NominateNodeReward{
		tx: ndau.NominateNodeReward{
			Random:   int64(random),
			Sequence: uint64(sequence),
		},
	}, nil
}

// ParseNominateNodeReward parses a string into a NominateNodeReward, if possible
func ParseNominateNodeReward(s string) (*NominateNodeReward, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseNominateNodeReward: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseNominateNodeReward: unmarshal")
	}
	trp, isTr := tx.(*ndau.NominateNodeReward)
	if !isTr {
		return nil, errors.New("ParseNominateNodeReward: transactable was not NominateNodeReward")
	}

	return &NominateNodeReward{tx: *trp}, nil
}

// ToString produces the b64 encoding of the bytes of the transaction
func (tx *NominateNodeReward) ToString() (string, error) {
	if tx == nil {
		return "", errors.New("nil nominatenodereward")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "nominatenodereward: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetRandom gets the random of the NominateNodeReward
//
// Returns `nil` if NominateNodeReward is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *NominateNodeReward) GetRandom() *int64 {
	if tx == nil {
		return nil
	}
	random := int64(tx.tx.Random)

	return &random
}

// GetSequence gets the sequence of the NominateNodeReward
//
// Returns `nil` if NominateNodeReward is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *NominateNodeReward) GetSequence() *int64 {
	if tx == nil {
		return nil
	}
	sequence := int64(tx.tx.Sequence)

	return &sequence
}

// GetNumSignatures gets the number of signatures of the NominateNodeReward
//
// If tx == nil, returns -1
func (tx *NominateNodeReward) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this NominateNodeReward
func (tx *NominateNodeReward) GetSignature(idx int) (*keyaddr.Signature, error) {
	if tx == nil {
		return nil, errors.New("nil nominatenodereward")
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

// SignableBytes returns the b64 encoding of the signable bytes of this nominatenodereward
func (tx *NominateNodeReward) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil nominatenodereward")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this nominatenodereward
func (tx *NominateNodeReward) AppendSignature(sig *keyaddr.Signature) error {
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

// Hash computes the hash of this nominatenodereward
func (tx *NominateNodeReward) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *NominateNodeReward) Name() string {
	if tx == nil {
		return ""
	}
	return "NominateNodeReward"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *NominateNodeReward) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
