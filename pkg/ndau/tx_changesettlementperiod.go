package ndau

import (
	"encoding/binary"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// SignableBytes implements Transactable
func (cep *ChangeSettlementPeriod) SignableBytes() []byte {
	bytes := make([]byte, 8+8, len(cep.Target.String())+8+8)
	binary.BigEndian.PutUint64(bytes[0:8], cep.Sequence)
	binary.BigEndian.PutUint64(bytes[8:16], uint64(cep.Period))
	bytes = append(bytes, []byte(cep.Target.String())...)
	return bytes
}

// NewChangeSettlementPeriod creates a new signed settlement period change
func NewChangeSettlementPeriod(
	target address.Address,
	newPeriod math.Duration,
	sequence uint64,
	privateTransferKey signature.PrivateKey,
) (ChangeSettlementPeriod, error) {
	cep := ChangeSettlementPeriod{
		Target:   target,
		Period:   newPeriod,
		Sequence: sequence,
	}
	sb := cep.SignableBytes()
	cep.Signature = privateTransferKey.Sign(sb)
	return cep, nil
}

// Validate implements metatx.Transactable
func (cep *ChangeSettlementPeriod) Validate(appI interface{}) (err error) {
	app := appI.(*App)

	if cep.Period < 0 {
		return errors.New("Negative settlement period")
	}

	acct := app.GetState().(*backing.State).Accounts[cep.Target.String()]

	if cep.Sequence <= acct.Sequence {
		return errors.New("Sequence too low")
	}

	if acct.TransferKey == nil {
		return errors.New("Target transfer key not set")
	}
	sb := cep.SignableBytes()
	if !acct.TransferKey.Verify(sb, cep.Signature) {
		return errors.New("Invalid message signature")
	}

	return nil
}

// Apply implements metatx.Transactable
func (cep *ChangeSettlementPeriod) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		acct := state.Accounts[cep.Target.String()]
		acct.UpdateSettlement(app.blockTime)
		acct.Sequence = cep.Sequence

		ca := app.blockTime.Add(acct.SettlementSettings.Period)
		acct.SettlementSettings.ChangesAt = &ca
		acct.SettlementSettings.Next = &cep.Period

		state.Accounts[cep.Target.String()] = acct
		return state, nil
	})
}
