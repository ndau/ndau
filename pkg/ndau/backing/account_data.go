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
//nomsify AccountData Stake Settlement SettlementSettings

// Stake keeps track of an account's staking information
type Stake struct {
	Point   math.Timestamp  `chain:"101,Stake_Point"`
	Address address.Address `chain:"102,Stake_Address"`
}

// Settlement tracks a single inbound transaction not yet settled
type Settlement struct {
	Qty math.Ndau `chain:"81,Settlement_Quantity"`
	// Expiry is when these funds are available to be sent
	Expiry math.Timestamp `chain:"82,Settlement_Expiry"`
}

// SettlementSettings tracks the settlement settings for outbound transactions
type SettlementSettings struct {
	Period    math.Duration   `json:"period" chain:"111,SettlementSettings_Period"`
	ChangesAt *math.Timestamp `json:"changesAt" chain:"112,SettlementSettings_ChangesAt"`
	Next      *math.Duration  `json:"next" chain:"113,SettlementSettings_Next"`
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
		SettlementSettings: SettlementSettings{
			Period: defaultSettlementPeriod,
		},
	}
}

// AccountData contains all the information the node needs to take action on a particular account.
//
// See the whitepaper: https://github.com/oneiro-ndev/whitepapers/blob/master/node_incentives/transactions.md#wallet-data
type AccountData struct {
	Balance             math.Ndau             `json:"balance" chain:"61,Acct_Balance"`
	ValidationKeys      []signature.PublicKey `json:"validationKeys" chain:"62,Acct_ValidationKeys"`
	ValidationScript    []byte                `json:"validationScript" chain:"69,Acct_ValidationScript"`
	RewardsTarget       *address.Address      `json:"rewardsTarget" chain:"63,Acct_RewardsTarget"`
	IncomingRewardsFrom []address.Address     `json:"incomingRewardsFrom" chain:"64,Acct_IncomingRewardsFrom"`
	DelegationNode      *address.Address      `json:"delegationNode" chain:"65,Acct_DelegationNode"`
	Lock                *Lock                 `json:"lock" chain:"."`
	Stake               *Stake                `json:"stake" chain:"."`
	StakeRules          []byte                `json:"stake_rules" chain:"75,Acct_StakeRules"`
	LastEAIUpdate       math.Timestamp        `json:"lastEAIUpdate" chain:"66,Acct_LastEAIUpdate"`
	LastWAAUpdate       math.Timestamp        `json:"lastWAAUpdate" chain:"67,Acct_LastWAAUpdate"`
	WeightedAverageAge  math.Duration         `json:"weightedAverageAge" chain:"68,Acct_WeightedAverageAge"`
	Sequence            uint64                `json:"sequence" chain:"71,Acct_Sequence"`
	Settlements         []Settlement          `json:"settlements" chain:"70,Acct_Settlements"`
	SettlementSettings  SettlementSettings    `json:"settlementSettings" chain:"."`
	CurrencySeatDate    *math.Timestamp       `json:"currencySeatDate" chain:"72,Acct_CurrencySeatDate"`
	Parent              *address.Address      `json:"parent" chain:"73,Acct_Parent"`
	Progenitor          *address.Address      `json:"progenitor" chain:"74,Acct_Progenitor"`
	UncreditedEAI       math.Ndau             `json:"-" msg:"-"` // exclude from serialization
}
