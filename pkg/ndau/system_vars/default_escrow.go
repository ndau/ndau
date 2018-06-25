package system_vars

import math "github.com/oneiro-ndev/ndaumath/pkg/types"

// DefaultEscrowDurationName is the name of the DefaultEscrowDuration system variable
const DefaultEscrowDurationName = "DefaultEscrowDuration"

// DefaultEscrowDuration is the system variable governing the default escrow duration
//
// It's a struct instead of a typedef so that it inherits methods
type DefaultEscrowDuration struct {
	math.Duration
}
