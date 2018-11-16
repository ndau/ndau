package query

// These constants define the endpoints at which the Tm RPC will forward requests
const (
	AccountEndpoint = "/account"
	SummaryEndpoint = "/summary"
	VersionEndpoint = "/version"
	SysvarsEndpoint = "/sysvars"
)

// AccountInfoFmt is used to set the format used by the Account endpoint in the
// Info field for whether the account exists. It is used with Sprintf on send
// and with Sscanf on receive.
const AccountInfoFmt = "acct exists: %t"
