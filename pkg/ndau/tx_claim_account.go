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
func NewClaimAccount(account address.Address, ownership signature.PublicKey, transferKeys []signature.PublicKey, ownerPrivate signature.PrivateKey) ClaimAccount {
	ca := ClaimAccount{
		Account:      account,
		Ownership:    ownership,
		TransferKeys: transferKeys,
	}
	ca.Signature = ownerPrivate.Sign(ca.SignableBytes())
	return ca
}

// SignableBytes returns the signable bytes of ClaimAccount
func (tx *ClaimAccount) SignableBytes() []byte {
	bnum := 0
	bnum += address.AddrLength
	bnum += tx.Ownership.Size()
	for _, key := range tx.TransferKeys {
		bnum += key.Size()
	}

	bytes := make([]byte, 0, bnum)

	bytes = append(bytes, tx.Account.String()...)
	bytes = append(bytes, tx.Ownership.Bytes()...)
	for _, key := range tx.TransferKeys {
		bytes = append(bytes, key.Bytes()...)
	}

	return bytes
}

// Validate returns nil if tx is valid, or an error
func (tx *ClaimAccount) Validate(appI interface{}) error {
	// we need to verify that the ownership key submitted actually generates
	// the address being claimed
	// get the address kind:
	_, err := address.Validate(tx.Account.String())
	if err != nil {
		return errors.Wrap(err, "Account address invalid")
	}
	kind := address.Kind(string(tx.Account.String()[2]))
	if !address.IsValidKind(kind) {
		return fmt.Errorf("Account has invalid address kind: %s", kind)
	}
	ownershipAddress, err := address.Generate(kind, tx.Ownership.Bytes())
	if err != nil {
		return errors.Wrap(err, "generating address for ownership key")
	}

	if tx.Account.String() != ownershipAddress.String() {
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

	app := appI.(*App)
	state := app.GetState().(*backing.State)

	// normally, we'd use GetValidAccount, but we can't do that here:
	// sequence validation is unusual, and we explicitly require no transfer
	// keys to be set
	acct, _ := state.GetAccount(
		tx.Account,
		app.blockTime,
	)

	if acct.Sequence > 0 {
		return errors.New("sequence must be 0 to claim account")
	}
	if len(acct.TransferKeys) > 0 {
		return errors.New("no transfer keys may be set to claim account")
	}

	return nil
}

// Apply applies this tx if no error occurs
func (tx *ClaimAccount) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		acct, _ := st.GetAccount(tx.Account, app.blockTime)
		acct.TransferKeys = tx.TransferKeys
		st.Accounts[tx.Account.String()] = acct

		return st, nil
	})
}
