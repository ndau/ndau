package mobile

// generated with github.com/oneiro-ndev/ndau/pkg/transactions.mobile/generator
// DO NOT EDIT

import (
	"encoding/base64"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/keyaddr"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau ChangeSettlementPeriod transaction
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

// ChangeSettlementPeriod is a mobile compatible wrapper for a ChangeSettlementPeriod transaction
type ChangeSettlementPeriod struct {
	tx ndau.ChangeSettlementPeriod
}

// NewChangeSettlementPeriod constructs a new unsigned ChangeSettlementPeriod transaction
func NewChangeSettlementPeriod(
	target *keyaddr.Address,
	period int64,
	sequence int64,
) (*ChangeSettlementPeriod, error) {
	if target == nil {
		return nil, errors.New("target must not be nil")
	}
	targetN, err := address.Validate(target.Address)
	if err != nil {
		return nil, errors.Wrap(err, "target")
	}

	return &ChangeSettlementPeriod{
		tx: ndau.ChangeSettlementPeriod{
			Target:   targetN,
			Period:   math.Duration(period),
			Sequence: uint64(sequence),
		},
	}, nil
}

// ParseChangeSettlementPeriod parses a string into a ChangeSettlementPeriod, if possible
func ParseChangeSettlementPeriod(s string) (*ChangeSettlementPeriod, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseChangeSettlementPeriod: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseChangeSettlementPeriod: unmarshal")
	}
	trp, isTr := tx.(*ndau.ChangeSettlementPeriod)
	if !isTr {
		return nil, errors.New("ParseChangeSettlementPeriod: transactable was not ChangeSettlementPeriod")
	}

	return &ChangeSettlementPeriod{tx: *trp}, nil
}

// ToString produces the b64 encoding of the bytes of the transaction
func (tx *ChangeSettlementPeriod) ToString() (string, error) {
	if tx == nil {
		return "", errors.New("nil changesettlementperiod")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "changesettlementperiod: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetTarget gets the target of the ChangeSettlementPeriod
//
// Returns `nil` if ChangeSettlementPeriod is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ChangeSettlementPeriod) GetTarget() *keyaddr.Address {
	if tx == nil {
		return nil
	}
	target := keyaddr.Address{Address: tx.tx.Target.String()}

	return &target
}

// GetPeriod gets the period of the ChangeSettlementPeriod
//
// Returns `nil` if ChangeSettlementPeriod is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ChangeSettlementPeriod) GetPeriod() *int64 {
	if tx == nil {
		return nil
	}
	period := int64(tx.tx.Period)

	return &period
}

// GetSequence gets the sequence of the ChangeSettlementPeriod
//
// Returns `nil` if ChangeSettlementPeriod is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *ChangeSettlementPeriod) GetSequence() *int64 {
	if tx == nil {
		return nil
	}
	sequence := int64(tx.tx.Sequence)

	return &sequence
}

// GetNumSignatures gets the number of signatures of the ChangeSettlementPeriod
//
// If tx == nil, returns -1
func (tx *ChangeSettlementPeriod) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this ChangeSettlementPeriod
func (tx *ChangeSettlementPeriod) GetSignature(idx int) (*keyaddr.Signature, error) {
	if tx == nil {
		return nil, errors.New("nil changesettlementperiod")
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

// SignableBytes returns the b64 encoding of the signable bytes of this changesettlementperiod
func (tx *ChangeSettlementPeriod) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil changesettlementperiod")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this changesettlementperiod
func (tx *ChangeSettlementPeriod) AppendSignature(sig *keyaddr.Signature) error {
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

// Hash computes the hash of this changesettlementperiod
func (tx *ChangeSettlementPeriod) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *ChangeSettlementPeriod) Name() string {
	if tx == nil {
		return ""
	}
	return "ChangeSettlementPeriod"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *ChangeSettlementPeriod) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
