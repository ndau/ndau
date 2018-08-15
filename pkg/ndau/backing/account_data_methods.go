package backing

import (
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

// MaxKeysInAccount is the maximum number of keys allowed to be associated
// with an account, and the maximum number of signatures allowed to be sent
// in a valid transaction.
const MaxKeysInAccount = 16

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

// ValidateSignatures returns `true` if signature quantity makes sense and
// every signature provided is valid given the provided data.
//
// It returns the validity of the signature set and a bitset. This bitset
// is a map: `1` elements are keys from `ad.TransferKeys` which validated a
// signature.
func (ad *AccountData) ValidateSignatures(data []byte, signatures []signature.Signature) (bool, *bitset256.Bitset256) {
	if len(signatures) < 1 || len(signatures) > MaxKeysInAccount {
		return false, nil
	}
	signatureSet := bitset256.New()

	// we could get fancy, making a map from each transfer key to its index,
	// using that to update the bitset, so that we could minimize the number
	// of test validations required. However, this would eliminate at most half
	// the field, causing us to check 128 signatures instead of 256. For these
	// values of N, I'm not sure that the work we'd save would actually pay for
	// the increase in setup cost. Instead, we're going to go with the simple
	// dumb solution: just check every key against every signature.
	allKeysValidate := true
	for _, signature := range signatures {
		foundValidatingKey := false
		for idx, key := range ad.TransferKeys {
			if key.Verify(data, signature) {
				foundValidatingKey = true
				signatureSet.Set(idx)
				break
			}
		}
		allKeysValidate = allKeysValidate && foundValidatingKey
	}
	return allKeysValidate, signatureSet
}
