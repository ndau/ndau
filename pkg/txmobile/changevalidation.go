package txmobile

// generated with github.com/oneiro-ndev/ndau/pkg/txmobile/generator
//

import (
	"encoding/base64"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau ChangeValidation transaction
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

// ChangeValidation is a mobile compatible wrapper for a ChangeValidation transaction
type ChangeValidation struct {
	tx ndau.ChangeValidation
}

// NewChangeValidation constructs a new unsigned ChangeValidation transaction
func NewChangeValidation(
	target string,
	validationscript string,
	sequence int64,
) (*ChangeValidation, error) {
	targetN, err := address.Validate(target)
	if err != nil {
		return nil, errors.Wrap(err, "target")
	}

	validationscriptN, err := base64.StdEncoding.DecodeString(validationscript)
	if err != nil {
		return nil, errors.Wrap(err, "validationscript")
	}

	return &ChangeValidation{
		tx: ndau.ChangeValidation{
			Target:           targetN,
			ValidationScript: validationscriptN,
			Sequence:         uint64(sequence),
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

// ToB64String produces the b64 encoding of the bytes of the transaction
func (tx *ChangeValidation) ToB64String() (string, error) {
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
// Returns a zero value if ChangeValidation is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ChangeValidation) GetTarget() string {
	if tx == nil {
		return ""
	}
	target := tx.tx.Target.String()

	return target
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
func (tx *ChangeValidation) GetNewKey(idx int) (string, error) {
	if tx == nil {
		return "", errors.New("nil changevalidation")
	}
	if idx < 0 || idx >= len(tx.tx.NewKeys) {
		return "", errors.New("invalid index")
	}
	newkey, err := tx.tx.NewKeys[idx].MarshalString()
	if err != nil {
		return "", errors.Wrap(err, "newkeys")
	}

	return newkey, nil
}

// AppendNewKey adds a newkey to the ChangeValidation
func (tx *ChangeValidation) AppendNewKey(newkey string) error {
	newkeysN, err := signature.ParsePublicKey(newkey)
	if err != nil {
		return errors.Wrap(err, "newkeys")
	}

	tx.tx.NewKeys = append(tx.tx.NewKeys, *newkeysN)

	return nil
}

// ClearNewKeys removes all newkeys from the ChangeValidation
func (tx *ChangeValidation) ClearNewKeys() {
	tx.tx.NewKeys = []signature.PublicKey{}
}

// GetValidationScript gets the validationscript of the ChangeValidation
//
// Returns a zero value if ChangeValidation is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ChangeValidation) GetValidationScript() string {
	if tx == nil {
		return ""
	}
	validationscript := base64.StdEncoding.EncodeToString(tx.tx.ValidationScript)

	return validationscript
}

// GetSequence gets the sequence of the ChangeValidation
//
// Returns a zero value if ChangeValidation is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ChangeValidation) GetSequence() int64 {
	if tx == nil {
		return 0
	}
	sequence := int64(tx.tx.Sequence)

	return sequence
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
func (tx *ChangeValidation) GetSignature(idx int) (string, error) {
	if tx == nil {
		return "", errors.New("nil changevalidation")
	}
	if idx < 0 || idx >= len(tx.tx.Signatures) {
		return "", errors.New("invalid index")
	}
	signature, err := tx.tx.Signatures[idx].MarshalString()
	if err != nil {
		return "", errors.Wrap(err, "signatures")
	}

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
func (tx *ChangeValidation) AppendSignature(sig string) error {
	if sig == "" {
		return errors.New("sig must not be blank")
	}
	sigS, err := signature.ParseSignature(sig)
	if err != nil {
		return errors.Wrap(err, "converting signature")
	}
	tx.tx.Signatures = append(tx.tx.Signatures, *sigS)
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
