package system_vars

import math "github.com/oneiro-ndev/ndaumath/pkg/types"

// DefaultSettlementDurationName is the name of the DefaultSettlementDuration system variable
const DefaultSettlementDurationName = "DefaultSettlementDuration"

// DefaultSettlementDuration is the system variable governing the default settlement duration
//
// It's a struct instead of a typedef so that it inherits methods
type DefaultSettlementDuration struct {
	math.Duration
}
