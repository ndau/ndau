package txmobile

// generated with github.com/oneiro-ndev/ndau/pkg/txmobile/generator
//

import (
	"encoding/base64"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau RecordPrice transaction
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

// RecordPrice is a mobile compatible wrapper for a RecordPrice transaction
type RecordPrice struct {
	tx ndau.RecordPrice
}

// NewRecordPrice constructs a new unsigned RecordPrice transaction
func NewRecordPrice(
	marketprice,
	sequence int64,
) (*RecordPrice, error) {
	return &RecordPrice{
		tx: ndau.RecordPrice{
			MarketPrice: marketprice,
			Sequence:    uint64(sequence),
		},
	}, nil
}

// ParseRecordPrice parses a string into a RecordPrice, if possible
func ParseRecordPrice(s string) (*RecordPrice, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseRecordPrice: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseRecordPrice: unmarshal")
	}
	trp, isTr := tx.(*ndau.RecordPrice)
	if !isTr {
		return nil, errors.New("ParseRecordPrice: transactable was not RecordPrice")
	}

	return &RecordPrice{tx: *trp}, nil
}

// ToB64String produces the b64 encoding of the bytes of the transaction
func (tx *RecordPrice) ToB64String() (string, error) {
	if tx == nil {
		return "", errors.New("nil recordprice")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "recordprice: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetMarketPrice gets the marketprice of the RecordPrice
//
// Returns a zero value if RecordPrice is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *RecordPrice) GetMarketPrice() {
	if tx == nil {
		return *new()
	}
	marketprice := tx.tx.MarketPrice

	return marketprice
}

// GetSequence gets the sequence of the RecordPrice
//
// Returns a zero value if RecordPrice is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *RecordPrice) GetSequence() int64 {
	if tx == nil {
		return 0
	}
	sequence := int64(tx.tx.Sequence)

	return sequence
}

// GetNumSignatures gets the number of signatures of the RecordPrice
//
// If tx == nil, returns -1
func (tx *RecordPrice) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this RecordPrice
func (tx *RecordPrice) GetSignature(idx int) (string, error) {
	if tx == nil {
		return "", errors.New("nil recordprice")
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

// SignableBytes returns the b64 encoding of the signable bytes of this recordprice
func (tx *RecordPrice) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil recordprice")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this recordprice
func (tx *RecordPrice) AppendSignature(sig string) error {
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

// Hash computes the hash of this recordprice
func (tx *RecordPrice) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *RecordPrice) Name() string {
	if tx == nil {
		return ""
	}
	return "RecordPrice"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *RecordPrice) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
