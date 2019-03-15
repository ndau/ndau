package txmobile

// generated with github.com/oneiro-ndev/ndau/pkg/txmobile/generator
//

import (
	"encoding/base64"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau ClaimChildAccount transaction
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

// ClaimChildAccount is a mobile compatible wrapper for a ClaimChildAccount transaction
type ClaimChildAccount struct {
	tx ndau.ClaimChildAccount
}

// NewClaimChildAccount constructs a new unsigned ClaimChildAccount transaction
func NewClaimChildAccount(
	target string,
	child string,
	childownership string,
	childsignature string,
	childsettlementperiod int64,
	childvalidationscript string,
	sequence int64,
) (*ClaimChildAccount, error) {
	targetN, err := address.Validate(target)
	if err != nil {
		return nil, errors.Wrap(err, "target")
	}

	childN, err := address.Validate(child)
	if err != nil {
		return nil, errors.Wrap(err, "child")
	}

	childownershipN, err := signature.ParsePublicKey(childownership)
	if err != nil {
		return nil, errors.Wrap(err, "childownership")
	}

	childsignatureN, err := signature.ParseSignature(childsignature)
	if err != nil {
		return nil, errors.Wrap(err, "childsignature")
	}

	childvalidationscriptN, err := base64.StdEncoding.DecodeString(childvalidationscript)
	if err != nil {
		return nil, errors.Wrap(err, "childvalidationscript")
	}

	return &ClaimChildAccount{
		tx: ndau.ClaimChildAccount{
			Target:                targetN,
			Child:                 childN,
			ChildOwnership:        *childownershipN,
			ChildSignature:        *childsignatureN,
			ChildSettlementPeriod: math.Duration(childsettlementperiod),
			ChildValidationScript: childvalidationscriptN,
			Sequence:              uint64(sequence),
		},
	}, nil
}

// ParseClaimChildAccount parses a string into a ClaimChildAccount, if possible
func ParseClaimChildAccount(s string) (*ClaimChildAccount, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseClaimChildAccount: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseClaimChildAccount: unmarshal")
	}
	trp, isTr := tx.(*ndau.ClaimChildAccount)
	if !isTr {
		return nil, errors.New("ParseClaimChildAccount: transactable was not ClaimChildAccount")
	}

	return &ClaimChildAccount{tx: *trp}, nil
}

// ToB64String produces the b64 encoding of the bytes of the transaction
func (tx *ClaimChildAccount) ToB64String() (string, error) {
	if tx == nil {
		return "", errors.New("nil claimchildaccount")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "claimchildaccount: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetTarget gets the target of the ClaimChildAccount
//
// Returns a zero value if ClaimChildAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimChildAccount) GetTarget() string {
	if tx == nil {
		return ""
	}
	target := tx.tx.Target.String()

	return target
}

// GetChild gets the child of the ClaimChildAccount
//
// Returns a zero value if ClaimChildAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimChildAccount) GetChild() string {
	if tx == nil {
		return ""
	}
	child := tx.tx.Child.String()

	return child
}

// GetChildOwnership gets the childownership of the ClaimChildAccount
//
// Returns a zero value if ClaimChildAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimChildAccount) GetChildOwnership() (string, error) {
	if tx == nil {
		return "", errors.New("nil ClaimChildAccount")
	}
	childownership, err := tx.tx.ChildOwnership.MarshalString()
	if err != nil {
		return "", errors.Wrap(err, "childownership")
	}

	return childownership, nil
}

// GetChildSignature gets the childsignature of the ClaimChildAccount
//
// Returns a zero value if ClaimChildAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimChildAccount) GetChildSignature() (string, error) {
	if tx == nil {
		return "", errors.New("nil ClaimChildAccount")
	}
	childsignature, err := tx.tx.ChildSignature.MarshalString()
	if err != nil {
		return "", errors.Wrap(err, "childsignature")
	}

	return childsignature, nil
}

// GetChildSettlementPeriod gets the childsettlementperiod of the ClaimChildAccount
//
// Returns a zero value if ClaimChildAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimChildAccount) GetChildSettlementPeriod() int64 {
	if tx == nil {
		return 0
	}
	childsettlementperiod := int64(tx.tx.ChildSettlementPeriod)

	return childsettlementperiod
}

// GetNumChildValidationKeys gets the number of childvalidationkeys of the ClaimChildAccount
//
// If tx == nil, returns -1
func (tx *ClaimChildAccount) GetNumChildValidationKeys() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.ChildValidationKeys)
}

// GetChildValidationKey gets a particular childvalidationkey from this ClaimChildAccount
func (tx *ClaimChildAccount) GetChildValidationKey(idx int) (string, error) {
	if tx == nil {
		return "", errors.New("nil claimchildaccount")
	}
	if idx < 0 || idx >= len(tx.tx.ChildValidationKeys) {
		return "", errors.New("invalid index")
	}
	childvalidationkey, err := tx.tx.ChildValidationKeys[idx].MarshalString()
	if err != nil {
		return "", errors.Wrap(err, "childvalidationkeys")
	}

	return childvalidationkey, nil
}

// AppendChildValidationKey adds a childvalidationkey to the ClaimChildAccount
func (tx *ClaimChildAccount) AppendChildValidationKey(childvalidationkey string) error {
	childvalidationkeysN, err := signature.ParsePublicKey(childvalidationkey)
	if err != nil {
		return errors.Wrap(err, "childvalidationkeys")
	}

	tx.tx.ChildValidationKeys = append(tx.tx.ChildValidationKeys, *childvalidationkeysN)

	return nil
}

// ClearChildValidationKeys removes all childvalidationkeys from the ClaimChildAccount
func (tx *ClaimChildAccount) ClearChildValidationKeys() {
	tx.tx.ChildValidationKeys = []signature.PublicKey{}
}

// GetChildValidationScript gets the childvalidationscript of the ClaimChildAccount
//
// Returns a zero value if ClaimChildAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimChildAccount) GetChildValidationScript() string {
	if tx == nil {
		return ""
	}
	childvalidationscript := base64.StdEncoding.EncodeToString(tx.tx.ChildValidationScript)

	return childvalidationscript
}

// GetSequence gets the sequence of the ClaimChildAccount
//
// Returns a zero value if ClaimChildAccount is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ClaimChildAccount) GetSequence() int64 {
	if tx == nil {
		return 0
	}
	sequence := int64(tx.tx.Sequence)

	return sequence
}

// GetNumSignatures gets the number of signatures of the ClaimChildAccount
//
// If tx == nil, returns -1
func (tx *ClaimChildAccount) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this ClaimChildAccount
func (tx *ClaimChildAccount) GetSignature(idx int) (string, error) {
	if tx == nil {
		return "", errors.New("nil claimchildaccount")
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

// SignableBytes returns the b64 encoding of the signable bytes of this claimchildaccount
func (tx *ClaimChildAccount) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil claimchildaccount")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this claimchildaccount
func (tx *ClaimChildAccount) AppendSignature(sig string) error {
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

// Hash computes the hash of this claimchildaccount
func (tx *ClaimChildAccount) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *ClaimChildAccount) Name() string {
	if tx == nil {
		return ""
	}
	return "ClaimChildAccount"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *ClaimChildAccount) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
