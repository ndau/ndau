package mobile

// generated with github.com/oneiro-ndev/ndau/pkg/transactions.mobile/generator
// DO NOT EDIT

import (
	"encoding/base64"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/b32"
	"github.com/oneiro-ndev/ndaumath/pkg/keyaddr"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau changevalidation transaction
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

// go build fails when there are unused imports, but we can't know a priori
// which imports will actually be used in a particular transaction.
// Therefore, let's force use of the frequent offenders
var (
	_ address.Address
	_ = b32.NdauAlphabet
	_ math.Ndau
	_ signature.Signature
)

// ChangeValidation is a mobile compatible wrapper for a ChangeValidation transaction
type ChangeValidation struct {
	tx ndau.ChangeValidation
}

// NewChangeValidation constructs a new unsigned ChangeValidation transaction
func NewChangeValidation(
	target *keyaddr.Address,
	newkeys []*keyaddr.Key,
	validationscript string,
	sequence int64,
) (*ChangeValidation, error) {
	if target == nil {
		return nil, errors.New("target must not be nil")
	}
	targetN, err := address.Validate(target.Address)
	if err != nil { return nil, errors.Wrap(err, "target") }

	
	if newkeys == nil {
		return nil, errors.New("newkeys must not be nil")
	}
	newkeysS := make([]signature.PublicKey, len(newkeys))
	for idx := range newkeys {
		newkeysN, err := newkeys[idx].ToPublicKey()
		newkeysS[idx] = newkeysN
		if err != nil { return nil, errors.Wrap(err, "newkeys") }

	}
	validationscriptN, err := base64.StdEncoding.DecodeString(validationscript)
	if err != nil { return nil, errors.Wrap(err, "validationscript") }

	
	return &ChangeValidation{
		tx: ndau.ChangeValidation{
			Target: targetN,
			NewKeys: newkeysS,
			ValidationScript: validationscriptN,
			Sequence: uint64(sequence),
		},
	}, nil
}

// ParseChangeValidation parses a string into a ChangeValidation, if possible
func ParseChangeValidation(s string) (*ChangeValidation, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseChangeValidation: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseChangeValidation: unmarshal")
	}
	trp, isTr := tx.(*ndau.ChangeValidation)
	if !isTr {
		return nil, errors.New("ParseChangeValidation: transactable was not ChangeValidation")
	}

	return &ChangeValidation{tx: *trp}, nil
}

// ToString produces the b64 encoding of the bytes of the transaction
func (tx *ChangeValidation) ToString() (string, error) {
	if tx == nil {
		return "", errors.New("nil changevalidation")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "changevalidation: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}



// GetTarget gets the target of the ChangeValidation
//
// Returns `nil` if ChangeValidation is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ChangeValidation) GetTarget() *keyaddr.Address {
	if tx == nil {
		return nil
	}
	target := keyaddr.Address{Address: tx.tx.Target.String()}
	
	return &target
}

// GetNumNewKeys gets the number of newkeys of the ChangeValidation
//
// If tx == nil, returns -1
func (tx *ChangeValidation) GetNumNewKeys() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.NewKeys)
}

// GetNewKey gets a particular newkey from this ChangeValidation
func (tx *ChangeValidation) GetNewKey(idx int) (*keyaddr.Key, error) {
	if tx == nil {
		return nil, errors.New("nil changevalidation")
	}
	if idx < 0 || idx >= len(tx.tx.NewKeys) {
		return nil, errors.New("invalid index")
	}
	newkey, err := keyaddr.KeyFromPublic(tx.tx.NewKeys[idx])
	if err != nil { return nil, errors.Wrap(err, "newkeys") }

	return newkey, nil
}


// GetValidationScript gets the validationscript of the ChangeValidation
//
// Returns `nil` if ChangeValidation is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ChangeValidation) GetValidationScript() *string {
	if tx == nil {
		return nil
	}
	validationscript := base64.StdEncoding.EncodeToString(tx.tx.ValidationScript)
	
	return &validationscript
}

// GetSequence gets the sequence of the ChangeValidation
//
// Returns `nil` if ChangeValidation is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ChangeValidation) GetSequence() *int64 {
	if tx == nil {
		return nil
	}
	sequence := int64(tx.tx.Sequence)
	
	return &sequence
}

// GetNumSignatures gets the number of signatures of the ChangeValidation
//
// If tx == nil, returns -1
func (tx *ChangeValidation) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this ChangeValidation
func (tx *ChangeValidation) GetSignature(idx int) (*keyaddr.Signature, error) {
	if tx == nil {
		return nil, errors.New("nil changevalidation")
	}
	if idx < 0 || idx >= len(tx.tx.Signatures) {
		return nil, errors.New("invalid index")
	}
	signature, err := keyaddr.SignatureFrom(tx.tx.Signatures[idx])
	if err != nil { return nil, errors.Wrap(err, "signatures") }

	return signature, nil
}


// SignableBytes returns the b64 encoding of the signable bytes of this changevalidation
func (tx *ChangeValidation) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil changevalidation")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this changevalidation
func (tx *ChangeValidation) AppendSignature(sig *keyaddr.Signature) error {
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

// Hash computes the hash of this changevalidation
func (tx *ChangeValidation) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *ChangeValidation) Name() string {
	if tx == nil {
		return ""
	}
	return "ChangeValidation"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *ChangeValidation) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
