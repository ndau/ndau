package system_vars

// MinDurationBetweenNodeRewardNominationsName names the minimum duration
// permitted between node rewards nominations
//
// The system variable named by this must have the value math.Duration
const MinDurationBetweenNodeRewardNominationsName = "MinDurationBetweenNodeRewardNominations"

// NominateNodeRewardAddressName is the name of the NominateNodeRewardAddress system variable
//
// The value contained in this system variable must be of type address.Address
const NominateNodeRewardAddressName = "NominateNodeRewardAddress"

// NominateNodeRewardOwnershipName is the name of the public ownership key
const NominateNodeRewardOwnershipName = "NominateNodeRewardOwnership"

// NominateNodeRewardOwnershipPrivateName is the name of the public ownership key
const NominateNodeRewardOwnershipPrivateName = "NominateNodeRewardOwnershipPrivate"
