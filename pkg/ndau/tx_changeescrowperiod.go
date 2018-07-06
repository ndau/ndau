package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta.app/meta.state"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

func (cep *ChangeEscrowPeriod) signableBytes() ([]byte, error) {
	bytes := []byte(cep.Target.String())
	return cep.Period.MarshalMsg(bytes)
}

// NewChangeEscrowPeriod creates a new signed escrow period change
func NewChangeEscrowPeriod(
	target address.Address,
	newPeriod math.Duration,
	privateTransferKey signature.PrivateKey,
) (ChangeEscrowPeriod, error) {
	cep := ChangeEscrowPeriod{
		Target: target,
		Period: newPeriod,
	}
	sb, err := cep.signableBytes()
	if err != nil {
		return cep, err
	}
	cep.Signature = privateTransferKey.Sign(sb)
	return cep, nil
}

// Validate implements metatx.Transactable
func (cep *ChangeEscrowPeriod) Validate(appI interface{}) (err error) {
	app := appI.(*App)

	if cep.Period < 0 {
		return errors.New("Negative escrow period")
	}

	acct := app.GetState().(*backing.State).Accounts[cep.Target.String()]

	if acct.TransferKey == nil {
		return errors.New("Target transfer key not set")
	}
	sb, err := cep.signableBytes()
	if err != nil {
		return err
	}
	if !acct.TransferKey.Verify(sb, cep.Signature) {
		return errors.New("Invalid message signature")
	}

	return nil
}

// Apply implements metatx.Transactable
func (cep *ChangeEscrowPeriod) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		acct := state.Accounts[cep.Target.String()]
		acct.UpdateEscrow(app.blockTime)

		ca := app.blockTime.Add(acct.EscrowSettings.Duration)
		acct.EscrowSettings.ChangesAt = &ca
		acct.EscrowSettings.Next = &cep.Period

		state.Accounts[cep.Target.String()] = acct
		return state, nil
	})
}
