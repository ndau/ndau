package backing

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// generate msgp interface implementations for AccountData and supporting structs
// we can't generate the streaming interfaces, unfortunately, because the
// signature.* types don't implement those
//go:generate msgp -io=0

// generate noms marshaler implementations for appropriate types
//go:generate go run $GOPATH/src/github.com/oneiro-ndev/generator/cmd/nomsify $GOPATH/src/github.com/oneiro-ndev/ndau/pkg/ndau/backing
//go:generate find $GOPATH/src/github.com/oneiro-ndev/ndau/pkg/ndau/backing -name "*noms_gen*.go" -maxdepth 1 -exec goimports -w {} ;
//nomsify AccountData RecourseSettings

// RecourseSettings tracks the settlement settings for outbound transactions
type RecourseSettings struct {
	Period    math.Duration   `json:"period" chain:"111,RecourseSettings_Period"`
	ChangesAt *math.Timestamp `json:"changes_at" chain:"112,RecourseSettings_ChangesAt"`
	Next      *math.Duration  `json:"next" chain:"113,RecourseSettings_Next"`
}

// NewAccountData creates a new AccountData struct
//
// The zero value of AccountData is not useful, because AccountData needs
// to have non-zero values for LastEAIUpdate and LastWAAUpdate if its EAI
// and WAA calculations are to be accurate.
//
// Unfortunately, go being go, we can't require that this method is used,
// but we can provide it to make it easier to do the right thing.
func NewAccountData(blockTime math.Timestamp, defaultSettlementPeriod math.Duration) AccountData {
	return AccountData{
		LastEAIUpdate: blockTime,
		LastWAAUpdate: blockTime,
		RecourseSettings: RecourseSettings{
			Period: defaultSettlementPeriod,
		},
	}
}

// AccountData contains all the information the node needs to take action on a particular account.
//
// See the whitepaper: https://github.com/oneiro-ndev/whitepapers/blob/master/node_incentives/transactions.md#wallet-data
type AccountData struct {
	Balance             math.Ndau                    `json:"balance" chain:"61,Acct_Balance"`
	ValidationKeys      []signature.PublicKey        `json:"validationKeys" chain:"62,Acct_ValidationKeys"`
	ValidationScript    []byte                       `json:"validationScript" chain:"69,Acct_ValidationScript"`
	RewardsTarget       *address.Address             `json:"rewardsTarget" chain:"63,Acct_RewardsTarget"`
	IncomingRewardsFrom []address.Address            `json:"incomingRewardsFrom" chain:"64,Acct_IncomingRewardsFrom"`
	DelegationNode      *address.Address             `json:"delegationNode" chain:"65,Acct_DelegationNode"`
	Lock                *Lock                        `json:"lock" chain:"78"`
	LastEAIUpdate       math.Timestamp               `json:"lastEAIUpdate" chain:"66,Acct_LastEAIUpdate"`
	LastWAAUpdate       math.Timestamp               `json:"lastWAAUpdate" chain:"67,Acct_LastWAAUpdate"`
	WeightedAverageAge  math.Duration                `json:"weightedAverageAge" chain:"68,Acct_WeightedAverageAge"`
	Sequence            uint64                       `json:"sequence" chain:"71,Acct_Sequence"`
	StakeRules          *StakeRules                  `json:"stake_rules" chain:"79"`
	Costakers           map[string]map[string]uint64 `json:"costakers" chain:"76,Acct_Costakers"`
	Holds               []Hold                       `json:"holds" chain:"70,Acct_Holds"`
	RecourseSettings    RecourseSettings             `json:"recourseSettings" chain:"80"`
	CurrencySeatDate    *math.Timestamp              `json:"currencySeatDate" chain:"72,Acct_CurrencySeatDate"`
	Parent              *address.Address             `json:"parent" chain:"73,Acct_Parent"`
	Progenitor          *address.Address             `json:"progenitor" chain:"74,Acct_Progenitor"`
	UncreditedEAI       math.Ndau                    `json:"-" msg:"-"` // exclude from serialization
}
