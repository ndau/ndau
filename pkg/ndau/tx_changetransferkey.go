package ndau

import (
	"bytes"
	"fmt"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	"github.com/oneiro-ndev/signature/pkg/signature"

	"github.com/pkg/errors"
)

func (ct *ChangeTransferKey) signableBytes() []byte {
	target := []byte(ct.Target.String())
	newKey := ct.NewKey.Bytes()
	signingKey := ct.SigningKey.Bytes()
	keyKind := byte(ct.KeyKind)

	bytes := make(
		[]byte, 0,
		len(target)+len(newKey)+len(signingKey)+1,
	)
	bytes = append(bytes, target...)
	bytes = append(bytes, newKey...)
	bytes = append(bytes, signingKey...)
	bytes = append(bytes, keyKind)
	return bytes
}

// NewChangeTransferKey creates a new signed transfer key from its data and a private key
func NewChangeTransferKey(
	target address.Address,
	newKey, signingKey signature.PublicKey,
	keyKind SigningKeyKind,
	private signature.PrivateKey,
) ChangeTransferKey {
	ct := ChangeTransferKey{
		Target:     target,
		NewKey:     newKey,
		SigningKey: signingKey,
		KeyKind:    keyKind,
	}
	ct.Signature = private.Sign(ct.signableBytes())
	return ct
}

// IsValid implements metatx.Transactable
func (ct *ChangeTransferKey) IsValid(appI interface{}) (err error) {
	ct.Target, err = address.Validate(ct.Target.String())
	if err != nil {
		return
	}
	// validation for NewKey, SigningKey, Signature happens on deserialization

	app := appI.(*App)
	acct, hasAcct := app.GetState().(*backing.State).Accounts[ct.Target.String()]
	if !hasAcct {
		return fmt.Errorf("Target account %s does not exist", ct.Target)
	}

	// get the target address kind for later use:
	// we need to generate addresses for the signing key, to verify it matches
	// the actual ownership key, if used, and for the new transfer key,
	// to ensure it's not equal to the actual ownership key
	kind := address.Kind(string(ct.Target.String()[2]))
	if !address.IsValidKind(kind) {
		return fmt.Errorf("Target has invalid address kind: %s", kind)
	}

	// ensure the key kind checks out
	switch ct.KeyKind {
	case SigningKeyTransfer:
		if acct.TransferKey == nil {
			return fmt.Errorf("Invalid KeyKind: no current transfer key set")
		}
		if !bytes.Equal(ct.SigningKey.Bytes(), acct.TransferKey.Bytes()) {
			return fmt.Errorf("Signing key is not previous transfer key")
		}
	case SigningKeyOwnership:
		sigAddr, err := address.Generate(kind, ct.SigningKey.Bytes())
		if err != nil {
			return errors.Wrap(err, "Failed to generate address from signing key")
		}
		if sigAddr.String() != ct.Target.String() {
			return fmt.Errorf("Invalid signing key: key address does not match target address")
		}
	default:
		return errors.New("Unknown key kind")
	}

	// ensure the signature validates the signing key
	if !ct.SigningKey.Verify(ct.signableBytes(), ct.Signature) {
		return fmt.Errorf("Invalid signature")
	}

	// new transfer key must not equal existing transfer key
	if acct.TransferKey != nil && bytes.Equal(ct.NewKey.Bytes(), acct.TransferKey.Bytes()) {
		return fmt.Errorf("New transfer key must not equal existing transfer key")
	}

	// new transfer key must not equal ownership key
	ntAddr, err := address.Generate(kind, ct.NewKey.Bytes())
	if err != nil {
		return errors.Wrap(err, "Failed to generate address from new transfer key")
	}
	if ct.Target.String() == ntAddr.String() {
		return fmt.Errorf("New transfer key must not equal ownership key")
	}
	return
}

// Apply implements metatx.Transactable
func (ct *ChangeTransferKey) Apply(appI interface{}) error {
	return errors.New("not implemented")
}
