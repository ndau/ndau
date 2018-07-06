package system_vars

import "github.com/oneiro-ndev/signature/pkg/signature"

//go:generate msgp -io=0

// NominatorKeysName is the name of the NominatorKeys system variable
const NominatorKeysName = "NominatorKeys"

// NominatorKeys is the system variable holding the public keys which are authorized to sign NominateDelegate transactions
type NominatorKeys []signature.PublicKey
