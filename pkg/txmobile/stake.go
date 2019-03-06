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

// This file provides an interface to the Ndau Stake transaction
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

// Stake is a mobile compatible wrapper for a Stake transaction
type Stake struct {
	tx ndau.Stake
}

// NewStake constructs a new unsigned Stake transaction
func NewStake(
	target string,
	stakedaccount string,
	sequence int64,
) (*Stake, error) {
	targetN, err := address.Validate(target)
	if err != nil {
		return nil, errors.Wrap(err, "target")
	}

	stakedaccountN, err := address.Validate(stakedaccount)
	if err != nil {
		return nil, errors.Wrap(err, "stakedaccount")
	}

	return &Stake{
		tx: ndau.Stake{
			Target:        targetN,
			StakedAccount: stakedaccountN,
			Sequence:      uint64(sequence),
		},
	}, nil
}

// ParseStake parses a string into a Stake, if possible
func ParseStake(s string) (*Stake, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseStake: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseStake: unmarshal")
	}
	trp, isTr := tx.(*ndau.Stake)
	if !isTr {
		return nil, errors.New("ParseStake: transactable was not Stake")
	}

	return &Stake{tx: *trp}, nil
}

// ToB64String produces the b64 encoding of the bytes of the transaction
func (tx *Stake) ToB64String() (string, error) {
	if tx == nil {
		return "", errors.New("nil stake")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "stake: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetTarget gets the target of the Stake
//
// Returns a zero value if Stake is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Stake) GetTarget() string {
	if tx == nil {
		return ""
	}
	target := tx.tx.Target.String()

	return target
}

// GetStakedAccount gets the stakedaccount of the Stake
//
// Returns a zero value if Stake is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Stake) GetStakedAccount() string {
	if tx == nil {
		return ""
	}
	stakedaccount := tx.tx.StakedAccount.String()

	return stakedaccount
}

// GetSequence gets the sequence of the Stake
//
// Returns a zero value if Stake is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *Stake) GetSequence() int64 {
	if tx == nil {
		return 0
	}
	sequence := int64(tx.tx.Sequence)

	return sequence
}

// GetNumSignatures gets the number of signatures of the Stake
//
// If tx == nil, returns -1
func (tx *Stake) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this Stake
func (tx *Stake) GetSignature(idx int) (string, error) {
	if tx == nil {
		return "", errors.New("nil stake")
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

// SignableBytes returns the b64 encoding of the signable bytes of this stake
func (tx *Stake) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil stake")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this stake
func (tx *Stake) AppendSignature(sig string) error {
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

// Hash computes the hash of this stake
func (tx *Stake) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *Stake) Name() string {
	if tx == nil {
		return ""
	}
	return "Stake"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *Stake) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
