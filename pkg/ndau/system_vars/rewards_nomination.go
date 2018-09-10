package system_vars

import "github.com/oneiro-ndev/signature/pkg/signature"

// MinDurationBetweenNodeRewardNominationsName names the minimum duration
// permitted between node rewards nominations
//
// The system variable named by this must have the value math.Timestamp
const MinDurationBetweenNodeRewardNominationsName = "MinDurationBetweenNodeRewardNominations"

//go:generate msgp -io=0

// NominateNodeRewardKeysName is the name of the NominateNodeRewardKeys system variable
const NominateNodeRewardKeysName = "NominateNodeRewardKeys"

// NominateNodeRewardKeys is the system variable holding the public keys which are authorized to sign NominateNodeReward transactions
type NominateNodeRewardKeys []signature.PublicKey
