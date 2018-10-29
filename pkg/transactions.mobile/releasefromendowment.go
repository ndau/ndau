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

// This file provides an interface to the Ndau releasefromendowment transaction
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

// ReleaseFromEndowment is a mobile compatible wrapper for a ReleaseFromEndowment transaction
type ReleaseFromEndowment struct {
	tx ndau.ReleaseFromEndowment
}

// NewReleaseFromEndowment constructs a new unsigned ReleaseFromEndowment transaction
func NewReleaseFromEndowment(
	destination *keyaddr.Address,
	qty int64,
	sequence int64,
) (*ReleaseFromEndowment, error) {
	if destination == nil {
		return nil, errors.New("destination must not be nil")
	}
	destinationN, err := address.Validate(destination.Address)
	if err != nil { return nil, errors.Wrap(err, "destination") }

	
	return &ReleaseFromEndowment{
		tx: ndau.ReleaseFromEndowment{
			Destination: destinationN,
			Qty: math.Ndau(qty),
			Sequence: uint64(sequence),
		},
	}, nil
}

// ParseReleaseFromEndowment parses a string into a ReleaseFromEndowment, if possible
func ParseReleaseFromEndowment(s string) (*ReleaseFromEndowment, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseReleaseFromEndowment: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseReleaseFromEndowment: unmarshal")
	}
	trp, isTr := tx.(*ndau.ReleaseFromEndowment)
	if !isTr {
		return nil, errors.New("ParseReleaseFromEndowment: transactable was not ReleaseFromEndowment")
	}

	return &ReleaseFromEndowment{tx: *trp}, nil
}

// ToString produces the b64 encoding of the bytes of the transaction
func (tx *ReleaseFromEndowment) ToString() (string, error) {
	if tx == nil {
		return "", errors.New("nil releasefromendowment")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "releasefromendowment: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}



// GetDestination gets the destination of the ReleaseFromEndowment
//
// Returns `nil` if ReleaseFromEndowment is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ReleaseFromEndowment) GetDestination() *keyaddr.Address {
	if tx == nil {
		return nil
	}
	destination := keyaddr.Address{Address: tx.tx.Destination.String()}
	
	return &destination
}

// GetQty gets the qty of the ReleaseFromEndowment
//
// Returns `nil` if ReleaseFromEndowment is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ReleaseFromEndowment) GetQty() *int64 {
	if tx == nil {
		return nil
	}
	qty := int64(tx.tx.Qty)
	
	return &qty
}

// GetSequence gets the sequence of the ReleaseFromEndowment
//
// Returns `nil` if ReleaseFromEndowment is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ReleaseFromEndowment) GetSequence() *int64 {
	if tx == nil {
		return nil
	}
	sequence := int64(tx.tx.Sequence)
	
	return &sequence
}

// GetNumSignatures gets the number of signatures of the ReleaseFromEndowment
//
// If tx == nil, returns -1
func (tx *ReleaseFromEndowment) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this ReleaseFromEndowment
func (tx *ReleaseFromEndowment) GetSignature(idx int) (*keyaddr.Signature, error) {
	if tx == nil {
		return nil, errors.New("nil releasefromendowment")
	}
	if idx < 0 || idx >= len(tx.tx.Signatures) {
		return nil, errors.New("invalid index")
	}
	signature, err := keyaddr.SignatureFrom(tx.tx.Signatures[idx])
	if err != nil { return nil, errors.Wrap(err, "signatures") }

	return signature, nil
}


// SignableBytes returns the b64 encoding of the signable bytes of this releasefromendowment
func (tx *ReleaseFromEndowment) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil releasefromendowment")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this releasefromendowment
func (tx *ReleaseFromEndowment) AppendSignature(sig *keyaddr.Signature) error {
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

// Hash computes the hash of this releasefromendowment
func (tx *ReleaseFromEndowment) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *ReleaseFromEndowment) Name() string {
	if tx == nil {
		return ""
	}
	return "ReleaseFromEndowment"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *ReleaseFromEndowment) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
