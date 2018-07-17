package ndau

import (
	"bytes"
	"encoding/binary"
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta.app/meta.state"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/signature/pkg/signature"

	"github.com/pkg/errors"
)

func (ct *ChangeTransferKey) signableBytes() []byte {
	target := []byte(ct.Target.String())
	newKey := ct.NewKey.Bytes()
	signingKey := ct.SigningKey.Bytes()
	keyKind := byte(ct.KeyKind)

	bytes := make(
		[]byte, 8,
		len(target)+len(newKey)+len(signingKey)+1+8,
	)
	binary.BigEndian.PutUint64(bytes, ct.Sequence)
	bytes = append(bytes, target...)
	bytes = append(bytes, newKey...)
	bytes = append(bytes, signingKey...)
	bytes = append(bytes, keyKind)
	return bytes
}

// NewChangeTransferKey creates a new signed transfer key from its data and a private key
func NewChangeTransferKey(
	target address.Address,
	newKey signature.PublicKey,
	sequence uint64,
	keyKind SigningKeyKind,
	signingKey signature.PublicKey,
	private signature.PrivateKey,
) ChangeTransferKey {
	ct := ChangeTransferKey{
		Target:     target,
		NewKey:     newKey,
		SigningKey: signingKey,
		KeyKind:    keyKind,
		Sequence:   sequence,
	}
	ct.Signature = private.Sign(ct.signableBytes())
	return ct
}

// Validate implements metatx.Transactable
func (ct *ChangeTransferKey) Validate(appI interface{}) (err error) {
	ct.Target, err = address.Validate(ct.Target.String())
	if err != nil {
		return
	}
	// validation for NewKey, SigningKey, Signature happens on deserialization

	app := appI.(*App)
	acct := app.GetState().(*backing.State).Accounts[ct.Target.String()]

	if ct.Sequence <= acct.Sequence {
		return errors.New("Sequence too low")
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
	app := appI.(*App)
	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		ad, hasAd := state.Accounts[ct.Target.String()]
		if !hasAd {
			ad = backing.NewAccountData(app.blockTime)
		}
		ad.Sequence = ct.Sequence

		ad.TransferKey = &ct.NewKey

		// business rule: if we're changing with an ownership key, and the
		// current settlement period is zero, then we set the settlement period
		// to the default
		if ct.KeyKind == SigningKeyOwnership && ad.SettlementSettings.Period == 0 {
			defaultDuration := new(sv.DefaultSettlementDuration)
			err := app.System(sv.DefaultSettlementDurationName, defaultDuration)
			if err != nil {
				return state, errors.Wrap(err, "ChangeTransferKey.Apply get default settlement period")
			}
			ad.SettlementSettings.Period = defaultDuration.Duration
		}

		state.Accounts[ct.Target.String()] = ad
		return state, nil
	})
}
