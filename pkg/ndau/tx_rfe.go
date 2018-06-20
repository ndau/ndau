package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta.app/meta.state"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndaunode/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/signature/pkg/signature"

	"github.com/pkg/errors"
)

func (rfe *ReleaseFromEndowment) signableBytes() ([]byte, error) {
	bytes := make([]byte, 0, rfe.Destination.Msgsize()+rfe.Qty.Msgsize())
	bytes, err := rfe.Destination.MarshalMsg(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Destination")
	}
	bytes, err = rfe.Qty.MarshalMsg(bytes)
	err = errors.Wrap(err, "Qty")
	return bytes, err
}

// NewReleaseFromEndowment constructs a ReleaseFromEndowment transactable.
//
// The caller must ensure that `private` corresponds to a public key listed
// in the `ReleaseFromEndowmentKeys` system variable.
func NewReleaseFromEndowment(
	qty math.Ndau,
	destination address.Address,
	private signature.PrivateKey,
) (ReleaseFromEndowment, error) {
	rfe := ReleaseFromEndowment{
		Qty:         qty,
		Destination: destination,
	}
	sb, err := rfe.signableBytes()
	if err == nil {
		rfe.Signature = private.Sign(sb)
	}
	return rfe, err
}

// IsValid implements metatx.Transactable
func (rfe *ReleaseFromEndowment) IsValid(appI interface{}) error {
	app := appI.(*App)

	if rfe.Qty <= 0 {
		return errors.New("RFE qty may not be <= 0")
	}

	rfeKeys := make(sv.ReleaseFromEndowmentKeys, 0)
	err := app.System(sv.ReleaseFromEndowmentKeysName, &rfeKeys)
	if err != nil {
		return errors.Wrap(err, "RFE.IsValid app.System err")
	}
	sb, err := rfe.signableBytes()
	if err != nil {
		return errors.Wrap(err, "RFE.IsValid signableBytes")
	}
	valid := false
	for _, public := range rfeKeys {
		if public.Verify(sb, rfe.Signature) {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("No public key in %s verifies RFE signature", sv.ReleaseFromEndowmentKeysName)
	}

	return nil
}

// Apply implements metatx.Transactable
func (rfe *ReleaseFromEndowment) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		var err error
		state := stateI.(*backing.State)

		acct := state.Accounts[rfe.Destination.String()]
		acct.Balance, err = acct.Balance.Add(rfe.Qty)
		if err == nil {
			state.Accounts[rfe.Destination.String()] = acct
		}
		return state, err
	})
}
