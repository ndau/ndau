package ndau

// This file generated by txgen: https://github.com/oneiro-ndev/generator/pkg/txgen
// DO NOT EDIT

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// NewTransfer creates a new Transfer transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewTransfer(
	source address.Address,
	destination address.Address,
	qty math.Ndau,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *Transfer {
	tx := &Transfer{
		Source:      source,
		Destination: destination,
		Qty:         qty,
		Sequence:    sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewChangeValidation creates a new ChangeValidation transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewChangeValidation(
	target address.Address,
	newkeys []signature.PublicKey,
	validationscript []byte,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *ChangeValidation {
	tx := &ChangeValidation{
		Target:           target,
		NewKeys:          newkeys,
		ValidationScript: validationscript,
		Sequence:         sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewReleaseFromEndowment creates a new ReleaseFromEndowment transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewReleaseFromEndowment(
	destination address.Address,
	qty math.Ndau,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *ReleaseFromEndowment {
	tx := &ReleaseFromEndowment{
		Destination: destination,
		Qty:         qty,
		Sequence:    sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewChangeSettlementPeriod creates a new ChangeSettlementPeriod transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewChangeSettlementPeriod(
	target address.Address,
	period math.Duration,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *ChangeSettlementPeriod {
	tx := &ChangeSettlementPeriod{
		Target:   target,
		Period:   period,
		Sequence: sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewDelegate creates a new Delegate transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewDelegate(
	target address.Address,
	node address.Address,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *Delegate {
	tx := &Delegate{
		Target:   target,
		Node:     node,
		Sequence: sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewCreditEAI creates a new CreditEAI transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewCreditEAI(
	node address.Address,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *CreditEAI {
	tx := &CreditEAI{
		Node:     node,
		Sequence: sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewLock creates a new Lock transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewLock(
	target address.Address,
	period math.Duration,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *Lock {
	tx := &Lock{
		Target:   target,
		Period:   period,
		Sequence: sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewNotify creates a new Notify transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewNotify(
	target address.Address,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *Notify {
	tx := &Notify{
		Target:   target,
		Sequence: sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewSetRewardsDestination creates a new SetRewardsDestination transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewSetRewardsDestination(
	target address.Address,
	destination address.Address,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *SetRewardsDestination {
	tx := &SetRewardsDestination{
		Target:      target,
		Destination: destination,
		Sequence:    sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewClaimAccount creates a new ClaimAccount transactable
//
// If signing keys are present, the new transactable is signed with all of them
//
// Note: though this constructor supports an arbitrary number of signing keys,
// ClaimAccount supports only a single signature. Any keys set beyond the
// first are ignored.
func NewClaimAccount(
	target address.Address,
	ownership signature.PublicKey,
	validationkeys []signature.PublicKey,
	validationscript []byte,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *ClaimAccount {
	tx := &ClaimAccount{
		Target:           target,
		Ownership:        ownership,
		ValidationKeys:   validationkeys,
		ValidationScript: validationscript,
		Sequence:         sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		tx.Signature = signingKeys[0].Sign(bytes)
	}

	return tx
}

// NewStake creates a new Stake transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewStake(
	target address.Address,
	stakedaccount address.Address,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *Stake {
	tx := &Stake{
		Target:        target,
		StakedAccount: stakedaccount,
		Sequence:      sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewRegisterNode creates a new RegisterNode transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewRegisterNode(
	node address.Address,
	distributionscript []byte,
	rpcaddress string,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *RegisterNode {
	tx := &RegisterNode{
		Node:               node,
		DistributionScript: distributionscript,
		RPCAddress:         rpcaddress,
		Sequence:           sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewNominateNodeReward creates a new NominateNodeReward transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewNominateNodeReward(
	random int64,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *NominateNodeReward {
	tx := &NominateNodeReward{
		Random:   random,
		Sequence: sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewClaimNodeReward creates a new ClaimNodeReward transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewClaimNodeReward(
	node address.Address,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *ClaimNodeReward {
	tx := &ClaimNodeReward{
		Node:     node,
		Sequence: sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewTransferAndLock creates a new TransferAndLock transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewTransferAndLock(
	source address.Address,
	destination address.Address,
	qty math.Ndau,
	period math.Duration,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *TransferAndLock {
	tx := &TransferAndLock{
		Source:      source,
		Destination: destination,
		Qty:         qty,
		Period:      period,
		Sequence:    sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewCommandValidatorChange creates a new CommandValidatorChange transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewCommandValidatorChange(
	publickey []byte,
	power int64,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *CommandValidatorChange {
	tx := &CommandValidatorChange{
		PublicKey: publickey,
		Power:     power,
		Sequence:  sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewSidechainTx creates a new SidechainTx transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewSidechainTx(
	source address.Address,
	sidechainid byte,
	sidechainsignablebytes []byte,
	sidechainsignatures []signature.Signature,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *SidechainTx {
	tx := &SidechainTx{
		Source:                 source,
		SidechainID:            sidechainid,
		SidechainSignableBytes: sidechainsignablebytes,
		SidechainSignatures:    sidechainsignatures,
		Sequence:               sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewUnregisterNode creates a new UnregisterNode transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewUnregisterNode(
	node address.Address,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *UnregisterNode {
	tx := &UnregisterNode{
		Node:     node,
		Sequence: sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewUnstake creates a new Unstake transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewUnstake(
	target address.Address,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *Unstake {
	tx := &Unstake{
		Target:   target,
		Sequence: sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewIssue creates a new Issue transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewIssue(
	qty math.Ndau,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *Issue {
	tx := &Issue{
		Qty:      qty,
		Sequence: sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}

// NewClaimChildAccount creates a new ClaimChildAccount transactable
//
// If signing keys are present, the new transactable is signed with all of them
func NewClaimChildAccount(
	target address.Address,
	child address.Address,
	childownership signature.PublicKey,
	childsignature signature.Signature,
	childsettlementperiod math.Duration,
	childvalidationkeys []signature.PublicKey,
	childvalidationscript []byte,
	sequence uint64,
	signingKeys ...signature.PrivateKey,
) *ClaimChildAccount {
	tx := &ClaimChildAccount{
		Target:                target,
		Child:                 child,
		ChildOwnership:        childownership,
		ChildSignature:        childsignature,
		ChildSettlementPeriod: childsettlementperiod,
		ChildValidationKeys:   childvalidationkeys,
		ChildValidationScript: childvalidationscript,
		Sequence:              sequence,
	}
	if len(signingKeys) > 0 {
		bytes := tx.SignableBytes()
		for _, key := range signingKeys {
			tx.Signatures = append(tx.Signatures, key.Sign(bytes))
		}
	}

	return tx
}
