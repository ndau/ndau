# ndau Blockchain Transactions
The ndau blockchain is a permanent, immutable record of transactions submitted by ndau accounts. Transactions may transfer ndau from one account to another, modify an account's settings and properties, manage the circulating supply of ndau, pay for transactions on related blockchains, or manage the properties of an ndau network node.

## Transfers

### Transfer
```
type Transfer struct {
	Source      address.Address       `msg:"src" chain:"1,Tx_Source" json:"source"`
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination" json:"destination"`
	Qty         math.Ndau             `msg:"qty" chain:"11,Tx_Quantity" json:"qty"`
	Sequence    uint64                `msg:"seq" json:"sequence"`
	Signatures  []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Transfer and Lock
```
type TransferAndLock struct {
	Source      address.Address       `msg:"src" chain:"1,Tx_Source" json:"source"`
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination" json:"destination"`
	Qty         math.Ndau             `msg:"qty" chain:"11,Tx_Quantity" json:"qty"`
	Period      math.Duration         `msg:"per" chain:"21,Tx_Period" json:"period"`
	Sequence    uint64                `msg:"seq" json:"sequence"`
	Signatures  []signature.Signature `msg:"sig" json:"signatures"`
}
```
## Locking
### Lock
```
type Lock struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Period     math.Duration         `msg:"per" chain:"21,Tx_Period" json:"period"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Notify
```
type Notify struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Set Rewards Destination
```
type SetRewardsDestination struct {
	Target      address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination" json:"destination"`
	Sequence    uint64                `msg:"seq" json:"sequence"`
	Signatures  []signature.Signature `msg:"sig" json:"signatures"`
}
```
## Account Validation Rules and Keys
### Claim Account
```
type ClaimAccount struct {
	Target           address.Address       `msg:"tgt" json:"target"`
	Ownership        signature.PublicKey   `msg:"own" json:"ownership"`
	ValidationKeys   []signature.PublicKey `msg:"key" json:"validation_keys"`
	ValidationScript []byte                `msg:"val" json:"validation_script"`
	Sequence         uint64                `msg:"seq" json:"sequence"`
	Signature        signature.Signature   `msg:"sig" json:"signature"`
}
```
### Claim Child Account
When an account's initial validation rules are established, it may also be identified as a "child" account associated with an already-existing ndau account. A child account is in every way a normal ndau account: claiming an account as a child has two effects:
1. Common control or ownership is asserted and visible on the blockchain. An account holder creates new accounts that can be identified and verified as being associated with the parent account.
2. Special attributes of the parent account, if any, are inherited by the child account.
   
The only special attribute currently inherited is an account's identification as an approved, whitelisted exchange account. Each new exchange listing ndau will create one exchange account, and that account will be whitelisted in a chaos chain list. That account will be flagged as an authorized exchange account and will have the special properties of an exchange account: it will earn a constant 2% EAI rate and transfers from it will not be subject to SIB fees.

Exchanges need to create many additional accounts with these same special properties for their normal operation. Accounts claimed as children of a whitelisted exchange account will also be flagged as authorized exchange accounts.

A `ClaimChildAccount` transaction is submitted by the parent account claiming ownership. It must be signed by one of more of that account's validation keys and be permitted by its validation rules. It must also be signed by the child account's ownership key, as a `ClaimAccount` transaction would be, to establish that the claiming parent account also owns the child account.
```
type ClaimChildAccount struct {
	Target           address.Address       `msg:"tgt" json:"target"`
	Child            address.Address       `msg:"chd" json:"child"`
	Ownership        signature.PublicKey   `msg:"own" json:"ownership"`
	ValidationKeys   []signature.PublicKey `msg:"key" json:"validation_keys"`
	ValidationScript []byte                `msg:"val" json:"validation_script"`
	Sequence         uint64                `msg:"seq" json:"sequence"`
	Signature        signature.Signature   `msg:"sig" json:"signature"`
}
```
### Change Validation
```
type ChangeValidation struct {
	Target           address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	NewKeys          []signature.PublicKey `msg:"key" chain:"31,Tx_NewKeys" json:"new_keys"`
	ValidationScript []byte                `msg:"val" chain:"32,Tx_ValidationScript" json:"validation_script"`
	Sequence         uint64                `msg:"seq" json:"sequence"`
	Signatures       []signature.Signature `msg:"sig" json:"signatures"`
}
```
## Account Properties
### Change Settlement Period
```
type ChangeSettlementPeriod struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Period     math.Duration         `msg:"per" chain:"21,Tx_Period" json:"period"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Delegate
```
type Delegate struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node" json:"node"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```
## ndau Node Operations
### Credit EAI
```
type CreditEAI struct {
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node" json:"node"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Stake
```
type Stake struct {
	Target        address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	StakedAccount address.Address       `msg:"ska" chain:"5,Tx_StakedAccount" json:"staked_account"`
	Sequence      uint64                `msg:"seq" json:"sequence"`
	Signatures    []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Register Node
```
type RegisterNode struct {
	Node               address.Address       `msg:"nod" chain:"4,Tx_Node" json:"node"`
	DistributionScript []byte                `msg:"dis" chain:"33,Tx_DistributionScript" json:"distribution_script"`
	RPCAddress         string                `msg:"rpc" chain:"34,Tx_RPCAddress" json:"rpc_address"`
	Sequence           uint64                `msg:"seq" json:"sequence"`
	Signatures         []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Unregister Node
```
type UnregisterNode struct {
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node" json:"node"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Unstake
```
type Unstake struct {
	Target     address.Address       `msg:"tgt" chain:"3,Tx_Target" json:"target"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```
## ndau Network Operations
### Nominate Node Reward
```
type NominateNodeReward struct {
	Random     int64                 `msg:"rnd" chain:"41,Tx_Random" json:"random"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Claim Node Reward
```
type ClaimNodeReward struct {
	Node       address.Address       `msg:"nod" chain:"4,Tx_Node" json:"node"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Command Validator Change
```
type CommandValidatorChange struct {
	Power int64 `msg:"pow" chain:"17,Tx_Power" json:"power"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```
## Payments for Related Blockchains
### Sidechain Transaction
```
type SidechainTx struct {
	Source                 address.Address       `msg:"src" chain:"1,Tx_Source" json:"source"`
	SidechainID            byte                  `msg:"sch" chain:"42,Tx_SidechainID" json:"sidechain_id"`
	SidechainSignableBytes []byte                `msg:"ssb" chain:"43,Tx_SidechainSignableBytes" json:"sidechain_signable_bytes"`
	SidechainSignatures    []signature.Signature `msg:"ssg" chain:"44,Tx_SidechainSignatures" json:"sidechain_signatures"`
	Sequence               uint64                `msg:"seq" json:"sequence"`
	Signatures             []signature.Signature `msg:"sig" json:"signatures"`
}
```
## ndau Currency Issuance
### Release from Endowment
```
// A ReleaseFromEndowment transaction is used to release funds from the
// endowment into an individual account.
//
// It must be signed with the private key corresponding to one of the public
// keys listed in the system variable `ReleaseFromEndowmentKeys`.
type ReleaseFromEndowment struct {
	Destination address.Address       `msg:"dst" chain:"2,Tx_Destination" json:"destination"`
	Qty         math.Ndau             `msg:"qty" chain:"11,Tx_Quantity" json:"qty"`
	Sequence    uint64                `msg:"seq" json:"sequence"`
	Signatures  []signature.Signature `msg:"sig" json:"signatures"`
}
```
### Issue
```
type Issue struct {
	Qty        math.Ndau             `msg:"qty" chain:"11,Tx_Quantity" json:"qty"`
	Sequence   uint64                `msg:"seq" json:"sequence"`
	Signatures []signature.Signature `msg:"sig" json:"signatures"`
}
```