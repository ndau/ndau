package ndau

import (
	"encoding/binary"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

func intbytes(i uint64) []byte {
	ib := make([]byte, 8)
	binary.BigEndian.PutUint64(ib, i)
	return ib
}

// SignableBytes partially implements metatx.Transactable for Transfer
func (tx *Transfer) SignableBytes() []byte {
	blen := 0 + address.AddrLength + address.AddrLength + 8 + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Source.String())...)
	bytes = append(bytes, []byte(tx.Destination.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Qty))...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for ChangeValidation
func (tx *ChangeValidation) SignableBytes() []byte {
	blen := 0 + address.AddrLength + len(tx.ValidationScript) + 8
	for _, changevalidationNewKeys := range tx.NewKeys {
		blen += changevalidationNewKeys.Msgsize()
	}
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Target.String())...)
	for _, v := range tx.NewKeys {
		vBytes, err := v.MarshalMsg(nil)
		if err == nil {
			return nil
		}
		bytes = append(bytes, vBytes...)
	}
	bytes = append(bytes, []byte(tx.ValidationScript)...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for ReleaseFromEndowment
func (tx *ReleaseFromEndowment) SignableBytes() []byte {
	blen := 0 + address.AddrLength + 8 + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Destination.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Qty))...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for ChangeSettlementPeriod
func (tx *ChangeSettlementPeriod) SignableBytes() []byte {
	blen := 0 + address.AddrLength + 8 + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Target.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Period))...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for Delegate
func (tx *Delegate) SignableBytes() []byte {
	blen := 0 + address.AddrLength + address.AddrLength + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Target.String())...)
	bytes = append(bytes, []byte(tx.Node.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for CreditEAI
func (tx *CreditEAI) SignableBytes() []byte {
	blen := 0 + address.AddrLength + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Node.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for Lock
func (tx *Lock) SignableBytes() []byte {
	blen := 0 + address.AddrLength + 8 + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Target.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Period))...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for Notify
func (tx *Notify) SignableBytes() []byte {
	blen := 0 + address.AddrLength + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Target.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for SetRewardsDestination
func (tx *SetRewardsDestination) SignableBytes() []byte {
	blen := 0 + address.AddrLength + address.AddrLength + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Source.String())...)
	bytes = append(bytes, []byte(tx.Destination.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for ClaimAccount
func (tx *ClaimAccount) SignableBytes() []byte {
	blen := 0 + address.AddrLength + tx.Ownership.Msgsize() + len(tx.ValidationScript) + 8
	for _, claimaccountValidationKeys := range tx.ValidationKeys {
		blen += claimaccountValidationKeys.Msgsize()
	}
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Target.String())...)
	claimaccountOwnershipBytes, err := tx.Ownership.MarshalMsg(nil)
	if err == nil {
		return nil
	}
	bytes = append(bytes, claimaccountOwnershipBytes...)
	for _, v := range tx.ValidationKeys {
		vBytes, err := v.MarshalMsg(nil)
		if err == nil {
			return nil
		}
		bytes = append(bytes, vBytes...)
	}
	bytes = append(bytes, []byte(tx.ValidationScript)...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for Stake
func (tx *Stake) SignableBytes() []byte {
	blen := 0 + address.AddrLength + address.AddrLength + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Target.String())...)
	bytes = append(bytes, []byte(tx.Node.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for RegisterNode
func (tx *RegisterNode) SignableBytes() []byte {
	blen := 0 + address.AddrLength + len(tx.DistributionScript) + len(tx.RPCAddress) + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Node.String())...)
	bytes = append(bytes, []byte(tx.DistributionScript)...)
	bytes = append(bytes, []byte(tx.RPCAddress)...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for NominateNodeReward
func (tx *NominateNodeReward) SignableBytes() []byte {
	blen := 0 + 8 + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, intbytes(uint64(tx.Random))...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for ClaimNodeReward
func (tx *ClaimNodeReward) SignableBytes() []byte {
	blen := 0 + address.AddrLength + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Node.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for TransferAndLock
func (tx *TransferAndLock) SignableBytes() []byte {
	blen := 0 + address.AddrLength + address.AddrLength + 8 + 8 + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Source.String())...)
	bytes = append(bytes, []byte(tx.Destination.String())...)
	bytes = append(bytes, intbytes(uint64(tx.Qty))...)
	bytes = append(bytes, intbytes(uint64(tx.Period))...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for CommandValidatorChange
func (tx *CommandValidatorChange) SignableBytes() []byte {
	blen := 0 + len(tx.PublicKey) + 8 + 8
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.PublicKey)...)
	bytes = append(bytes, intbytes(uint64(tx.Power))...)
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}

// SignableBytes partially implements metatx.Transactable for SidechainTx
func (tx *SidechainTx) SignableBytes() []byte {
	blen := 0 + address.AddrLength + 1 + len(tx.SidechainSignableBytes) + 8
	for _, sidechaintxSidechainSignatures := range tx.SidechainSignatures {
		blen += sidechaintxSidechainSignatures.Msgsize()
	}
	bytes := make([]byte, 0, blen)

	bytes = append(bytes, []byte(tx.Source.String())...)
	bytes = append(bytes, []byte{tx.SidechainID}...)
	bytes = append(bytes, []byte(tx.SidechainSignableBytes)...)
	for _, v := range tx.SidechainSignatures {
		vBytes, err := v.MarshalMsg(nil)
		if err == nil {
			return nil
		}
		bytes = append(bytes, vBytes...)
	}
	bytes = append(bytes, intbytes(uint64(tx.Sequence))...)

	return bytes
}
