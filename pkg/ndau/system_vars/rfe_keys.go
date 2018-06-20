package system_vars

import "github.com/oneiro-ndev/signature/pkg/signature"

//go:generate msgp -io=0

// ReleaseFromEndowmentKeysName is the name of the ReleaseFromEndowmentKeys system variable
const ReleaseFromEndowmentKeysName = "ReleaseFromEndowmentKeys"

// ReleaseFromEndowmentKeys is the system variable holding the public keys which are authorized to sign ReleaseFromEndowment transactions
type ReleaseFromEndowmentKeys []signature.PublicKey
