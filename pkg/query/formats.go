package query

// These constants define the format strings which controls the information in the Info field of the relevant queries
const (
	AccountInfoFmt           = "acct exists: %t"
	PrevalidateInfoFmt       = "estimated tx fee: %d napu; estimated sib: %d napu"
	SidechainTxExistsInfoFmt = "sidechain tx paid for and validated: %t"
)
