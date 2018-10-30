package mobile

// generated with github.com/oneiro-ndev/ndau/pkg/transactions.mobile/generator
// DO NOT EDIT

import (
	"encoding/base64"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/keyaddr"
	"github.com/pkg/errors"
)

// This file provides an interface to the Ndau registernode transaction
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

// RegisterNode is a mobile compatible wrapper for a RegisterNode transaction
type RegisterNode struct {
	tx ndau.RegisterNode
}

// NewRegisterNode constructs a new unsigned RegisterNode transaction
func NewRegisterNode(
	node *keyaddr.Address,
	distributionscript string,
	rpcaddress string,
	sequence int64,
) (*RegisterNode, error) {
	if node == nil {
		return nil, errors.New("node must not be nil")
	}
	nodeN, err := address.Validate(node.Address)
	if err != nil {
		return nil, errors.Wrap(err, "node")
	}

	distributionscriptN, err := base64.StdEncoding.DecodeString(distributionscript)
	if err != nil {
		return nil, errors.Wrap(err, "distributionscript")
	}

	return &RegisterNode{
		tx: ndau.RegisterNode{
			Node:               nodeN,
			DistributionScript: distributionscriptN,
			RPCAddress:         rpcaddress,
			Sequence:           uint64(sequence),
		},
	}, nil
}

// ParseRegisterNode parses a string into a RegisterNode, if possible
func ParseRegisterNode(s string) (*RegisterNode, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, errors.Wrap(err, "ParseRegisterNode: b64-decode")
	}
	tx, err := metatx.Unmarshal(bytes, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "ParseRegisterNode: unmarshal")
	}
	trp, isTr := tx.(*ndau.RegisterNode)
	if !isTr {
		return nil, errors.New("ParseRegisterNode: transactable was not RegisterNode")
	}

	return &RegisterNode{tx: *trp}, nil
}

// ToString produces the b64 encoding of the bytes of the transaction
func (tx *RegisterNode) ToString() (string, error) {
	if tx == nil {
		return "", errors.New("nil registernode")
	}
	bytes, err := metatx.Marshal(&tx.tx, ndau.TxIDs)
	if err != nil {
		return "", errors.Wrap(err, "registernode: marshalling bytes")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GetNode gets the node of the RegisterNode
//
// Returns `nil` if RegisterNode is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *RegisterNode) GetNode() *keyaddr.Address {
	if tx == nil {
		return nil
	}
	node := keyaddr.Address{Address: tx.tx.Node.String()}

	return &node
}

// GetDistributionScript gets the distributionscript of the RegisterNode
//
// Returns `nil` if RegisterNode is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *RegisterNode) GetDistributionScript() *string {
	if tx == nil {
		return nil
	}
	distributionscript := base64.StdEncoding.EncodeToString(tx.tx.DistributionScript)

	return &distributionscript
}

// GetRPCAddress gets the rpcaddress of the RegisterNode
//
// Returns `nil` if RegisterNode is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *RegisterNode) GetRPCAddress() *string {
	if tx == nil {
		return nil
	}
	rpcaddress := tx.tx.RPCAddress

	return &rpcaddress
}

// GetSequence gets the sequence of the RegisterNode
//
// Returns `nil` if RegisterNode is `nil` or if native conversion is fallible and
// conversion failed.
func (tx *RegisterNode) GetSequence() *int64 {
	if tx == nil {
		return nil
	}
	sequence := int64(tx.tx.Sequence)

	return &sequence
}

// GetNumSignatures gets the number of signatures of the RegisterNode
//
// If tx == nil, returns -1
func (tx *RegisterNode) GetNumSignatures() int {
	if tx == nil {
		return -1
	}
	return len(tx.tx.Signatures)
}

// GetSignature gets a particular signature from this RegisterNode
func (tx *RegisterNode) GetSignature(idx int) (*keyaddr.Signature, error) {
	if tx == nil {
		return nil, errors.New("nil registernode")
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

// SignableBytes returns the b64 encoding of the signable bytes of this registernode
func (tx *RegisterNode) SignableBytes() (string, error) {
	if tx == nil {
		return "", errors.New("nil registernode")
	}
	return base64.StdEncoding.EncodeToString(tx.tx.SignableBytes()), nil
}

// AppendSignature appends a signature to this registernode
func (tx *RegisterNode) AppendSignature(sig *keyaddr.Signature) error {
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

// Hash computes the hash of this registernode
func (tx *RegisterNode) Hash() string {
	if tx == nil {
		return ""
	}
	return metatx.Hash(&tx.tx)
}

// Name returns the name of this transactable
func (tx *RegisterNode) Name() string {
	if tx == nil {
		return ""
	}
	return "RegisterNode"
}

// TxID returns the transaction id of this transactable
//
// Returns -2 if the transactable is nil, or -1 if the transactable is unknown.
func (tx *RegisterNode) TxID() int {
	if tx == nil {
		return -2
	}
	id, err := metatx.TxIDOf(&tx.tx, ndau.TxIDs)
	if err != nil {
		return -1
	}
	return int(id)
}
