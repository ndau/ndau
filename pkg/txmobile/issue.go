package txmobile

// generated with github.com/oneiro-ndev/ndau/pkg/txmobile/generator
// DO NOT EDIT

import (
	"encoding/base64"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau Issue transaction
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

// Issue is a mobile compatible wrapper for a Issue transaction
type Issue struct {
	tx ndau.Issue
}

// NewIssue constructs a new unsigned Issue transaction
func NewIssue(
	qty int64,
	sequence int64,
) (*Issue, error) {
	return &Issue{
		tx: ndau.Issue{
			Qty:      math.Ndau(qty),
			Sequence: uint64(sequence),
		},
	}, nil
}

// ParseIssue parses a string into a Issue, if possible
func ParseIssue(s string) (*Issue, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseIssue: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseIssue: unmarshal")
	}
	trp, isTr := tx.(*ndau.Issue)
	if !isTr {
		return nil, errors.New("ParseIssue: transactable was not Issue")
	}

	return &Issue{tx: *trp}, nil
}

// ToB64String produces the b64 encoding of the bytes of the transaction
func (tx *Issue) ToB64String() (string, error) {
	if tx == nil {
		return "", errors.New("nil issue")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "issue: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetQty gets the qty of the Issue
//
// Returns a zero value if Issue is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Issue) GetQty() int64 {
	if tx == nil {
		return 0
	}
	qty := int64(tx.tx.Qty)

	return qty
}

// GetSequence gets the sequence of the Issue
//
// Returns a zero value if Issue is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Issue) GetSequence() int64 {
	if tx == nil {
		return 0
	}
	sequence := int64(tx.tx.Sequence)

	return sequence
}

// GetNumSignatures gets the number of signatures of the Issue
//
// If tx == nil, returns -1
func (tx *Issue) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this Issue
func (tx *Issue) GetSignature(idx int) (string, error) {
	if tx == nil {
		return "", errors.New("nil issue")
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

// SignableBytes returns the b64 encoding of the signable bytes of this issue
func (tx *Issue) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil issue")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this issue
func (tx *Issue) AppendSignature(sig string) error {
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

// Hash computes the hash of this issue
func (tx *Issue) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *Issue) Name() string {
	if tx == nil {
		return ""
	}
	return "Issue"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *Issue) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
