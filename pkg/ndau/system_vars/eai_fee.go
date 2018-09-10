package system_vars

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

//go:generate msgp -io=0

// EAIFeeTableName names the EAI fee table
//
// The system variable of this name must have the type EAIFeeTable
const EAIFeeTableName = "EAIFeeTable"

// EAIFeeTable is a list of EAI fees and their destinations
type EAIFeeTable []EAIFee

// An EAIFee is a fee applied to accrued EAI when crediting.
//
// The fee is listed as Ndau; the listed value is multiplied by the number
// of Ndau actually earned as EAI.
//
// The fee is credited to the account at the listed address. If the destination
// is nil, it is considered to be a node reward, and is tracked in internal
// state instead of going into an account.
type EAIFee struct {
	Fee math.Ndau
	To  *address.Address
}
