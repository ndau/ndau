package txmobile

// generated with github.com/oneiro-ndev/ndau/pkg/txmobile/generator
//

import (
	"encoding/base64"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau Delegate transaction
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

// Delegate is a mobile compatible wrapper for a Delegate transaction
type Delegate struct {
	tx ndau.Delegate
}

// NewDelegate constructs a new unsigned Delegate transaction
func NewDelegate(
	target string,
	node string,
	sequence int64,
) (*Delegate, error) {
	targetN, err := address.Validate(target)
	if err != nil {
		return nil, errors.Wrap(err, "target")
	}

	nodeN, err := address.Validate(node)
	if err != nil {
		return nil, errors.Wrap(err, "node")
	}

	return &Delegate{
		tx: ndau.Delegate{
			Target:   targetN,
			Node:     nodeN,
			Sequence: uint64(sequence),
		},
	}, nil
}

// ParseDelegate parses a string into a Delegate, if possible
func ParseDelegate(s string) (*Delegate, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseDelegate: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseDelegate: unmarshal")
	}
	trp, isTr := tx.(*ndau.Delegate)
	if !isTr {
		return nil, errors.New("ParseDelegate: transactable was not Delegate")
	}

	return &Delegate{tx: *trp}, nil
}

// ToB64String produces the b64 encoding of the bytes of the transaction
func (tx *Delegate) ToB64String() (string, error) {
	if tx == nil {
		return "", errors.New("nil delegate")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "delegate: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetTarget gets the target of the Delegate
//
// Returns a zero value if Delegate is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Delegate) GetTarget() string {
	if tx == nil {
		return ""
	}
	target := tx.tx.Target.String()

	return target
}

// GetNode gets the node of the Delegate
//
// Returns a zero value if Delegate is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Delegate) GetNode() string {
	if tx == nil {
		return ""
	}
	node := tx.tx.Node.String()

	return node
}

// GetSequence gets the sequence of the Delegate
//
// Returns a zero value if Delegate is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Delegate) GetSequence() int64 {
	if tx == nil {
		return 0
	}
	sequence := int64(tx.tx.Sequence)

	return sequence
}

// GetNumSignatures gets the number of signatures of the Delegate
//
// If tx == nil, returns -1
func (tx *Delegate) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this Delegate
func (tx *Delegate) GetSignature(idx int) (string, error) {
	if tx == nil {
		return "", errors.New("nil delegate")
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

// SignableBytes returns the b64 encoding of the signable bytes of this delegate
func (tx *Delegate) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil delegate")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this delegate
func (tx *Delegate) AppendSignature(sig string) error {
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

// Hash computes the hash of this delegate
func (tx *Delegate) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *Delegate) Name() string {
	if tx == nil {
		return ""
	}
	return "Delegate"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *Delegate) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
