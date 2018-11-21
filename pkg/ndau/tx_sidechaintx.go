package ndau

import (
	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

// NewSidechainTx creates a new signed sidechaintx transactable
func NewSidechainTx(
	source address.Address,
	sidechain byte,
	txid metatx.TxID,
	txsize uint32,
	txhash string,
	sidechainSignatures []signature.Signature,
	seq uint64,
	keys []signature.PrivateKey,
) (*SidechainTx, error) {
	tx := &SidechainTx{
		Source:              source,
		SidechainID:         sidechain,
		TxID:                txid,
		TxSize:              txsize,
		TxHash:              txhash,
		SidechainSignatures: sidechainSignatures,
		Sequence:            seq,
	}
	bytes := tx.SignableBytes()
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(bytes))
	}

	return tx, nil
}

// SignableBytes implements Transactable
func (tx *SidechainTx) SignableBytes() []byte {
	bytes := make([]byte, 0, tx.Msgsize())
	bytes = append(bytes, tx.Source.String()...)
	bytes = append(bytes, tx.SidechainID, byte(tx.TxID))
	bytes = appendUint64(bytes, uint64(tx.TxSize))
	bytes = append(bytes, tx.TxHash...)
	for _, sig := range tx.SidechainSignatures {
		bytes = append(bytes, sig.Bytes()...)
	}
	bytes = appendUint64(bytes, tx.Sequence)
	return bytes
}

// Validate satisfies metatx.Transactable
func (tx *SidechainTx) Validate(appInt interface{}) error {
	app := appInt.(*App)

	// two things happen here:
	// 1. transaction fee calculated and compared to balance in source acct
	// 2. source acct validation script applied to tx contents
	_, _, _, err := app.getTxAccount(tx)
	return err
}

// Apply satisfies metatx.Transactable
func (tx *SidechainTx) Apply(appInt interface{}) error {
	app := appInt.(*App)
	return app.applyTxDetails(tx)
}

// GetSource implements sourcer
func (tx *SidechainTx) GetSource(*App) (address.Address, error) {
	return tx.Source, nil
}

// GetSequence implements sequencer
func (tx *SidechainTx) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *SidechainTx) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *SidechainTx) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
