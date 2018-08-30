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
	transferKeys []signature.PublicKey,
	validationScript []byte,
	sequence uint64,
	ownerPrivate signature.PrivateKey,
) ClaimAccount {
	ca := ClaimAccount{
		Target:           account,
		Ownership:        ownership,
		TransferKeys:     transferKeys,
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
	for _, key := range tx.TransferKeys {
		bcnt += key.Size()
	}
	bcnt += len(tx.ValidationScript)
	bcnt += 8 // sequence

	bytes := make([]byte, 0, bcnt)

	bytes = append(bytes, tx.Target.String()...)
	bytes = append(bytes, tx.Ownership.Bytes()...)
	for _, key := range tx.TransferKeys {
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
	if len(tx.TransferKeys) < 1 || len(tx.TransferKeys) > backing.MaxKeysInAccount {
		return fmt.Errorf("Expect between 1 and %d transfer keys; got %d", backing.MaxKeysInAccount, len(tx.TransferKeys))
	}

	// no transfer key may be equal to the ownership key
	for _, tk := range tx.TransferKeys {
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
	state := app.GetState().(*backing.State)

	// normally, we'd use GetValidAccount, but we can't do that here:
	// we have unusual requirements
	acct, _ := state.GetAccount(
		tx.Target,
		app.blockTime,
	)

	if tx.Sequence <= acct.Sequence {
		return errors.New("sequence is too low")
	}

	if len(acct.TransferKeys) > 1 {
		return errors.New("claim account is not valid if there are 2 or more transfer keys")
	}

	return nil
}

// Apply applies this tx if no error occurs
func (tx *ClaimAccount) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		acct, _ := st.GetAccount(tx.Target, app.blockTime)
		acct.TransferKeys = tx.TransferKeys
		acct.ValidationScript = tx.ValidationScript
		st.Accounts[tx.Target.String()] = acct
		acct.Sequence = tx.Sequence

		return st, nil
	})
}
