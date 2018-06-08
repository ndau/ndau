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
	if ad.Lock.NotifiedOn == nil {
		return true
	}
	if ad.Lock.NotifiedOn.Add(ad.Lock.NoticePeriod).Compare(blockTime) <= 0 {
		return true
	}
	ad.Lock = nil
	return false
}

// IsNotified is true when this account has been notified but has not yet unlocked.
func (ad *AccountData) IsNotified(blockTime math.Timestamp) bool {
	if ad.Lock == nil || ad.Lock.NotifiedOn == nil {
		return false
	}
	if ad.Lock.NotifiedOn.Add(ad.Lock.NoticePeriod).Compare(blockTime) <= 0 {
		return true
	}
	ad.Lock = nil
	return false
}
