package backing

import (
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
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

// UpdateSettlements settles funds whose settlement periods have expired
func (ad *AccountData) UpdateSettlements(blockTime math.Timestamp) {
	newSettlements := make([]Hold, 0, len(ad.Holds))
	for _, settlement := range ad.Holds {
		if settlement.Expiry == nil || settlement.Expiry.Compare(blockTime) > 0 {
			// blockTime > settlement.Expiry
			newSettlements = append(newSettlements, settlement)
		}
	}
	ad.Holds = newSettlements

	// true if there exists a pending change which is less than or equal to the block time
	if ad.RecourseSettings.ChangesAt != nil && blockTime.Compare(*ad.RecourseSettings.ChangesAt) >= 0 {
		ad.RecourseSettings.Period = *ad.RecourseSettings.Next
		ad.RecourseSettings.ChangesAt = nil
		ad.RecourseSettings.Next = nil
	}
}

// AvailableBalance computes the balance available
//
// Normally one would wish to call UpdateSettlements shortly before calling this
func (ad AccountData) AvailableBalance() (math.Ndau, error) {
	held := ad.HoldSum()
	var err error
	balance, err := ad.Balance.Sub(held)
	if err != nil {
		return math.Ndau(0), errors.Wrap(err, "available balance")

	}
	return balance, nil
}

// ValidateSignatures returns `true` if signature quantity makes sense and
// every signature provided is valid given the provided data.
//
// It returns the validity of the signature set and a bitset.
//
// The bitset is a map whose semantics differ according to whether or not
// validation succeeded.
//
// If validation was successful, `1` elements are keys from `ad.ValidationKeys`
// which validated a signature.
//
// If validation was not successful, `1` elements are elements from `signatures`
// which failed to validate.
func (ad *AccountData) ValidateSignatures(data []byte, signatures []signature.Signature) (bool, *bitset256.Bitset256) {
	signatureSet := bitset256.New()
	invalidSignatures := bitset256.New()

	if len(signatures) < 1 || len(signatures) > MaxKeysInAccount {
		return false, signatureSet
	}

	// we could get fancy, making a map from each transfer key to its index,
	// using that to update the bitset, so that we could minimize the number
	// of test validations required. However, this would eliminate at most half
	// the field, causing us to check 128 signatures instead of 256. For these
	// values of N, I'm not sure that the work we'd save would actually pay for
	// the increase in setup cost. Instead, we're going to go with the simple
	// dumb solution: just check every signature against every key that has
	// not already been used.
	allKeysValidate := true
	for sidx, signature := range signatures {
		foundValidatingKey := false
		for idx, key := range ad.ValidationKeys {
			// don't attempt to verify keys we've already verified
			if !signatureSet.Get(byte(idx)) {
				if key.Verify(data, signature) {
					foundValidatingKey = true
					signatureSet.Set(byte(idx))
					break
				}
			}
		}
		if !foundValidatingKey {
			invalidSignatures.Set(byte(sidx))
		}
		allKeysValidate = allKeysValidate && foundValidatingKey
	}
	// If everything validated but the signatureSet doesn't have as many bits set as
	// there were signatures, then we must have had duplicates, which is bad.
	valid := allKeysValidate && signatureSet.Count() == len(signatures)
	if valid {
		return valid, signatureSet
	}
	return valid, invalidSignatures
}

// UpdateCurrencySeat sets the account's currency seat status appropriately given
// its balance. It is safe to call repeatedly; it's smart enough not to change
// the state inappropriately.
func (ad *AccountData) UpdateCurrencySeat(blockTime math.Timestamp) {
	const currencySeatQty = 1000 * constants.NapuPerNdau

	if ad.CurrencySeatDate != nil && ad.Balance < currencySeatQty {
		ad.CurrencySeatDate = nil
	} else if ad.CurrencySeatDate == nil && ad.Balance >= currencySeatQty {
		ad.CurrencySeatDate = &blockTime
	}
}
