package generator

import (
	"encoding/base64"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/b32"
	"github.com/oneiro-ndev/ndaumath/pkg/keyaddr"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// Proof-of-concept: this file is a manual implementation so we can work out
// design questions, so that we can implement a generator which will emit
// something similar for each transaction.

// This package provides an interface to the Ndau transfer transaction for use in React and in particular react-native.
// It is built using the gomobile tool, so the API is constrained to particular types of parameters:
//
// * string
// * signed integer and floating point types
// * []byte
// * functions with specific restrictions
// * structs and interfaces consisting of only these types
//
// Unfortunately, react-native puts additional requirements that makes []byte particularly
// challenging to use. So what we are going to do is use a base-64 encoding of []byte to convert
// it to a string and pass the array of bytes back and forth that way.
//
// This is distinct from using base32 encoding (b32) in a signature; that's something we expect
// to be user-visible, so we're using a specific variant of base 32.

// This package, therefore, consists mainly of wrappers so that we don't have to modify our
// idiomatic Go code to conform to these requirements.

// // A Transfer is the fundamental transaction of the Ndau chain.
// type Transfer struct {
// 	Source      address.Address       `msg:"src" chain:"1,Tx_Source"`
// 	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination"`
// 	Qty         math.Ndau             `msg:"qty" chain:"11,Tx_Quantity"`
// 	Sequence    uint64                `msg:"seq"`
// 	Signatures  []signature.Signature `msg:"sig"`
// }

// A Transfer is the fundamental transaction of the Ndau chain.
type Transfer struct {
	tx ndau.Transfer
}

// NewTransfer constructs a new unsigned Transfer transaction
func NewTransfer(
	source *keyaddr.Address,
	dest *keyaddr.Address,
	qty int64,
	sequence int64,
) (*Transfer, error) {
	if source == nil {
		return nil, errors.New("source must not be nil")
	}
	sourceA, err := address.Validate(source.Address)
	if err != nil {
		return nil, errors.Wrap(err, "source")
	}
	if dest == nil {
		return nil, errors.New("dest must not be nil")
	}
	destA, err := address.Validate(dest.Address)
	if err != nil {
		return nil, errors.Wrap(err, "dest")
	}
	return &Transfer{
		tx: ndau.Transfer{
			Source:      sourceA,
			Destination: destA,
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
// We can't call it ToString because that conflicts with a native Java feature.
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

// GetSource gets the source of the transfer
func (tx *Transfer) GetSource() *keyaddr.Address {
	if tx == nil {
		return nil
	}
	return &keyaddr.Address{Address: tx.tx.Source.String()}
}

// GetDestination gets the destination of the transfer
func (tx *Transfer) GetDestination() *keyaddr.Address {
	if tx == nil {
		return nil
	}
	return &keyaddr.Address{Address: tx.tx.Destination.String()}
}

// GetQty gets the qty of the transfer
func (tx *Transfer) GetQty() *int64 {
	if tx == nil {
		return nil
	}
	q := int64(tx.tx.Qty)
	return &q
}

// GetSequence gets the sequence of the transfer
func (tx *Transfer) GetSequence() *int64 {
	if tx == nil {
		return nil
	}
	s := int64(tx.tx.Sequence)
	return &s
}

// GetNumSignatures gets the number of signatures of the transfer
func (tx *Transfer) GetNumSignatures() *int {
	if tx == nil {
		return nil
	}
	l := len(tx.tx.Signatures)
	return &l
}

// GetSignature gets a particular signature from this transfer
func (tx *Transfer) GetSignature(idx int) (*keyaddr.Signature, error) {
	if tx == nil {
		return nil, errors.New("nil transfer")
	}
	if idx < 0 || idx >= len(tx.tx.Signatures) {
		return nil, errors.New("invalid index")
	}
	sigB, err := tx.tx.Signatures[idx].Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshalling signature")
	}

	return &keyaddr.Signature{Signature: b32.Encode(sigB)}, nil
}

// SignableBytes returns the b64 encoding of the signable bytes of this transfer
func (tx *Transfer) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil transfer")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this transfer
func (tx *Transfer) AppendSignature(sig *keyaddr.Signature) error {
	if sig == nil {
		return errors.New("sig must not be nil")
	}
	sigB, err := b32.Decode(sig.Signature)
	if err != nil {
		return errors.Wrap(err, "decoding signature bytes")
	}
	sigS := new(signature.Signature)
	err = sigS.Unmarshal(sigB)
	if err != nil {
		return errors.Wrap(err, "unmarshalling signature")
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
	return metatx.NameOf(&tx.tx)
}

// TxID returns the transaction id of this transactable
func (tx *Transfer) TxID() int {
	if tx == nil {
		return 0
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
