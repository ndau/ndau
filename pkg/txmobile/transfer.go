package txmobile

// generated with github.com/oneiro-ndev/ndau/pkg/txmobile/generator
// DO NOT EDIT

import (
	"encoding/base64"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau Transfer transaction
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

// Transfer is a mobile compatible wrapper for a Transfer transaction
type Transfer struct {
	tx ndau.Transfer
}

// NewTransfer constructs a new unsigned Transfer transaction
func NewTransfer(
	source string,
	destination string,
	qty int64,
	sequence int64,
) (*Transfer, error) {
	sourceN, err := address.Validate(source)
	if err != nil {
		return nil, errors.Wrap(err, "source")
	}

	destinationN, err := address.Validate(destination)
	if err != nil {
		return nil, errors.Wrap(err, "destination")
	}

	return &Transfer{
		tx: ndau.Transfer{
			Source:      sourceN,
			Destination: destinationN,
			Qty:         math.Ndau(qty),
			Sequence:    uint64(sequence),
		},
	}, nil
}

// ParseTransfer parses a string into a Transfer, if possible
func ParseTransfer(s string) (*Transfer, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseTransfer: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseTransfer: unmarshal")
	}
	trp, isTr := tx.(*ndau.Transfer)
	if !isTr {
		return nil, errors.New("ParseTransfer: transactable was not Transfer")
	}

	return &Transfer{tx: *trp}, nil
}

// ToB64String produces the b64 encoding of the bytes of the transaction
func (tx *Transfer) ToB64String() (string, error) {
	if tx == nil {
		return "", errors.New("nil transfer")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "transfer: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetSource gets the source of the Transfer
//
// Returns a zero value if Transfer is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Transfer) GetSource() string {
	if tx == nil {
		return *new(string)
	}
	source := tx.tx.Source.String()

	return source
}

// GetDestination gets the destination of the Transfer
//
// Returns a zero value if Transfer is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Transfer) GetDestination() string {
	if tx == nil {
		return *new(string)
	}
	destination := tx.tx.Destination.String()

	return destination
}

// GetQty gets the qty of the Transfer
//
// Returns a zero value if Transfer is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Transfer) GetQty() int64 {
	if tx == nil {
		return *new(int64)
	}
	qty := int64(tx.tx.Qty)

	return qty
}

// GetSequence gets the sequence of the Transfer
//
// Returns a zero value if Transfer is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Transfer) GetSequence() int64 {
	if tx == nil {
		return *new(int64)
	}
	sequence := int64(tx.tx.Sequence)

	return sequence
}

// GetNumSignatures gets the number of signatures of the Transfer
//
// If tx == nil, returns -1
func (tx *Transfer) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this Transfer
func (tx *Transfer) GetSignature(idx int) (string, error) {
	if tx == nil {
		return *new(string), errors.New("nil transfer")
	}
	if idx < 0 || idx >= len(tx.tx.Signatures) {
		return *new(string), errors.New("invalid index")
	}
	signature, err := tx.tx.Signatures[idx].MarshalString()
	if err != nil {
		return *new(string), errors.Wrap(err, "signatures")
	}

	return signature, nil
}

// SignableBytes returns the b64 encoding of the signable bytes of this transfer
func (tx *Transfer) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil transfer")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this transfer
func (tx *Transfer) AppendSignature(sig string) error {
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

// Hash computes the hash of this transfer
func (tx *Transfer) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *Transfer) Name() string {
	if tx == nil {
		return ""
	}
	return "Transfer"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *Transfer) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
