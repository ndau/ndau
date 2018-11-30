package txmobile

// generated with github.com/oneiro-ndev/ndau/pkg/txmobile/generator
// DO NOT EDIT

import (
	"encoding/base64"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau ClaimAccount transaction
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

// ClaimAccount is a mobile compatible wrapper for a ClaimAccount transaction
type ClaimAccount struct {
	tx ndau.ClaimAccount
}

// NewClaimAccount constructs a new unsigned ClaimAccount transaction
func NewClaimAccount(
	target string,
	ownership string,
	validationscript string,
	sequence int64,
) (*ClaimAccount, error) {
	targetN, err := address.Validate(target)
	if err != nil {
		return nil, errors.Wrap(err, "target")
	}

	ownershipN, err := signature.ParsePublicKey(ownership)
	if err != nil {
		return nil, errors.Wrap(err, "ownership")
	}

	validationscriptN, err := base64.StdEncoding.DecodeString(validationscript)
	if err != nil {
		return nil, errors.Wrap(err, "validationscript")
	}

	return &ClaimAccount{
		tx: ndau.ClaimAccount{
			Target:           targetN,
			Ownership:        *ownershipN,
			ValidationScript: validationscriptN,
			Sequence:         uint64(sequence),
		},
	}, nil
}

// ParseClaimAccount parses a string into a ClaimAccount, if possible
func ParseClaimAccount(s string) (*ClaimAccount, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseClaimAccount: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseClaimAccount: unmarshal")
	}
	trp, isTr := tx.(*ndau.ClaimAccount)
	if !isTr {
		return nil, errors.New("ParseClaimAccount: transactable was not ClaimAccount")
	}

	return &ClaimAccount{tx: *trp}, nil
}

// ToB64String produces the b64 encoding of the bytes of the transaction
func (tx *ClaimAccount) ToB64String() (string, error) {
	if tx == nil {
		return "", errors.New("nil claimaccount")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "claimaccount: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetTarget gets the target of the ClaimAccount
//
// Returns a zero value if ClaimAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimAccount) GetTarget() string {
	if tx == nil {
		return ""
	}
	target := tx.tx.Target.String()

	return target
}

// GetOwnership gets the ownership of the ClaimAccount
//
// Returns a zero value if ClaimAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimAccount) GetOwnership() (string, error) {
	if tx == nil {
		return "", errors.New("nil ClaimAccount")
	}
	ownership, err := tx.tx.Ownership.MarshalString()
	if err != nil {
		return "", errors.Wrap(err, "ownership")
	}

	return ownership, nil
}

// GetNumValidationKeys gets the number of validationkeys of the ClaimAccount
//
// If tx == nil, returns -1
func (tx *ClaimAccount) GetNumValidationKeys() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.ValidationKeys)
}

// GetValidationKey gets a particular validationkey from this ClaimAccount
func (tx *ClaimAccount) GetValidationKey(idx int) (string, error) {
	if tx == nil {
		return "", errors.New("nil claimaccount")
	}
	if idx < 0 || idx >= len(tx.tx.ValidationKeys) {
		return "", errors.New("invalid index")
	}
	validationkey, err := tx.tx.ValidationKeys[idx].MarshalString()
	if err != nil {
		return "", errors.Wrap(err, "validationkeys")
	}

	return validationkey, nil
}

// AppendValidationKey adds a validationkey to the ClaimAccount
func (tx *ClaimAccount) AppendValidationKey(validationkey string) error {
	validationkeysN, err := signature.ParsePublicKey(validationkey)
	if err != nil {
		return errors.Wrap(err, "validationkeys")
	}

	tx.tx.ValidationKeys = append(tx.tx.ValidationKeys, *validationkeysN)

	return nil
}

// ClearValidationKeys removes all validationkeys from the ClaimAccount
func (tx *ClaimAccount) ClearValidationKeys() {
	tx.tx.ValidationKeys = []signature.PublicKey{}
}

// GetValidationScript gets the validationscript of the ClaimAccount
//
// Returns a zero value if ClaimAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimAccount) GetValidationScript() string {
	if tx == nil {
		return ""
	}
	validationscript := base64.StdEncoding.EncodeToString(tx.tx.ValidationScript)

	return validationscript
}

// GetSequence gets the sequence of the ClaimAccount
//
// Returns a zero value if ClaimAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimAccount) GetSequence() int64 {
	if tx == nil {
		return 0
	}
	sequence := int64(tx.tx.Sequence)

	return sequence
}

// GetSignature gets the signature of the ClaimAccount
//
// Returns a zero value if ClaimAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimAccount) GetSignature() (string, error) {
	if tx == nil {
		return "", errors.New("nil ClaimAccount")
	}
	signature, err := tx.tx.Signature.MarshalString()
	if err != nil {
		return "", errors.Wrap(err, "signature")
	}

	return signature, nil
}

// SignableBytes returns the b64 encoding of the signable bytes of this claimaccount
func (tx *ClaimAccount) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil claimaccount")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// Sign signs this claimaccount
func (tx *ClaimAccount) Sign(sig string) error {
	if sig == "" {
		return errors.New("sig must not be blank")
	}
	sigS, err := signature.ParseSignature(sig)
	if err != nil {
		return errors.Wrap(err, "converting signature")
	}
	tx.tx.Signature = *sigS
	return nil
}

// Hash computes the hash of this claimaccount
func (tx *ClaimAccount) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *ClaimAccount) Name() string {
	if tx == nil {
		return ""
	}
	return "ClaimAccount"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *ClaimAccount) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
