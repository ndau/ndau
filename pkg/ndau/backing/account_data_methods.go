package backing

import math "github.com/oneiro-ndev/ndaumath/pkg/types"

// IsLocked abstracts away the complexity of our lock model.
//
// If the account is locked and not notified, this returns true.
// If the account is notified and the notification period has not
// yet elapsed, this returns true.
// If the account is notified and the notification period has elapsed,
// this returns false and removes the lock.
func (ad *AccountData) IsLocked(blockTime math.Timestamp) bool {
	if ad.Lock == nil {
		return false
	}
	if ad.Lock.UnlocksOn == nil {
		return true
	}
	if ad.Lock.UnlocksOn.Compare(blockTime) > 0 {
		return true
	}
	ad.Lock = nil
	return false
}

// IsNotified is true when this account has been notified but has not yet unlocked.
func (ad *AccountData) IsNotified(blockTime math.Timestamp) bool {
	if ad.Lock == nil || ad.Lock.UnlocksOn == nil {
		return false
	}
	if ad.Lock.UnlocksOn.Compare(blockTime) > 0 {
		return true
	}
	ad.Lock = nil
	return false
}

// UpdateEscrow adds escrowed funds whose escrows have expired to the balance.
func (ad *AccountData) UpdateEscrow(blockTime math.Timestamp) {
	newEscrows := make([]Escrow, 0, len(ad.Escrows))
	for _, escrow := range ad.Escrows {
		if escrow.Expiry.Compare(blockTime) <= 0 {
			ad.Balance += escrow.Qty
		} else {
			newEscrows = append(newEscrows, escrow)
		}
	}
	ad.Escrows = newEscrows

	if ad.EscrowSettings.ChangesAt != nil && blockTime.Compare(*ad.EscrowSettings.ChangesAt) >= 0 {
		ad.EscrowSettings.Duration = *ad.EscrowSettings.Next
		ad.EscrowSettings.ChangesAt = nil
		ad.EscrowSettings.Next = nil
	}
}
