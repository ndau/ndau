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

// UpdateSettlement settles funds whose settlement periods have expired
func (ad *AccountData) UpdateSettlement(blockTime math.Timestamp) {
	newSettlements := make([]Settlement, 0, len(ad.Settlements))
	for _, settlement := range ad.Settlements {
		if settlement.Expiry.Compare(blockTime) <= 0 {
			ad.Balance += settlement.Qty
		} else {
			newSettlements = append(newSettlements, settlement)
		}
	}
	ad.Settlements = newSettlements

	// true if there exists a pending change which is less than or equal to the block time
	if ad.SettlementSettings.ChangesAt != nil && blockTime.Compare(*ad.SettlementSettings.ChangesAt) >= 0 {
		ad.SettlementSettings.Period = *ad.SettlementSettings.Next
		ad.SettlementSettings.ChangesAt = nil
		ad.SettlementSettings.Next = nil
	}
}
