package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewClaimAccount creates a ClaimAccount transaction
func NewClaimAccount(
	account address.Address,
	ownership signature.PublicKey,
	validationKeys []signature.PublicKey,
	validationScript []byte,
	sequence uint64,
	ownerPrivate signature.PrivateKey,
) ClaimAccount {
	ca := ClaimAccount{
		Target:           account,
		Ownership:        ownership,
		ValidationKeys:   validationKeys,
		ValidationScript: validationScript,
		Sequence:         sequence,
	}
	ca.Signature = ownerPrivate.Sign(ca.SignableBytes())
	return ca
}

// SignableBytes returns the signable bytes of ClaimAccount
func (tx *ClaimAccount) SignableBytes() []byte {
	bcnt := 0
	bcnt += address.AddrLength
	bcnt += tx.Ownership.Size()
	for _, key := range tx.ValidationKeys {
		bcnt += key.Size()
	}
	bcnt += len(tx.ValidationScript)
	bcnt += 8 // sequence

	bytes := make([]byte, 0, bcnt)

	bytes = append(bytes, tx.Target.String()...)
	bytes = append(bytes, tx.Ownership.Bytes()...)
	for _, key := range tx.ValidationKeys {
		bytes = append(bytes, key.Bytes()...)
	}
	bytes = append(bytes, tx.ValidationScript...)
	bytes = appendUint64(bytes, tx.Sequence)

	return bytes
}

// Validate returns nil if tx is valid, or an error
func (tx *ClaimAccount) Validate(appI interface{}) error {
	// we need to verify that the ownership key submitted actually generates
	// the address being claimed
	// get the address kind:
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err, "Account address invalid")
	}
	kind := address.Kind(string(tx.Target.String()[2]))
	if !address.IsValidKind(kind) {
		return fmt.Errorf("Account has invalid address kind: %s", kind)
	}
	ownershipAddress, err := address.Generate(kind, tx.Ownership.Bytes())
	if err != nil {
		return errors.Wrap(err, "generating address for ownership key")
	}

	if tx.Target.String() != ownershipAddress.String() {
		return errors.New("Ownership key and address do not match")
	}

	if !tx.Signature.Verify(tx.SignableBytes(), tx.Ownership) {
		return errors.New("Invalid ownership signature")
	}

	// business rule: there must be at least 1 and no more than a const
	// transfer keys set in this tx
	if len(tx.ValidationKeys) < 1 || len(tx.ValidationKeys) > backing.MaxKeysInAccount {
		return fmt.Errorf("Expect between 1 and %d transfer keys; got %d", backing.MaxKeysInAccount, len(tx.ValidationKeys))
	}

	// no transfer key may be equal to the ownership key
	for _, tk := range tx.ValidationKeys {
		tkAddress, err := address.Generate(kind, tk.Bytes())
		if err != nil {
			return errors.Wrap(err, "generating address for transfer key")
		}
		if tkAddress.String() == ownershipAddress.String() {
			return errors.New("Ownership key may not be used as a transfer key")
		}
	}

	if len(tx.ValidationScript) > 0 && !IsChaincode(tx.ValidationScript) {
		return errors.New("Validation script must be chaincode")
	}

	app := appI.(*App)

	acct, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if len(acct.ValidationKeys) > 1 {
		return errors.New("claim account is not valid if there are 2 or more transfer keys")
	}

	return nil
}

// Apply applies this tx if no error occurs
func (tx *ClaimAccount) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		acct, _ := st.GetAccount(tx.Target, app.blockTime)
		acct.ValidationKeys = tx.ValidationKeys
		acct.ValidationScript = tx.ValidationScript

		st.Accounts[tx.Target.String()] = acct

		return st, nil
	})
}

// GetSource implements sourcer
func (tx *ClaimAccount) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements sequencer
func (tx *ClaimAccount) GetSequence() uint64 {
	return tx.Sequence
}
