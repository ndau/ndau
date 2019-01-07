package ndau

import (
	"testing"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
)

func TestTransfer_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	transferSource, err := address.Validate("ndafm5p7wkpyubvk9crg4yg6bdmg3qavhhhu6wk4idxrzbs3")
	require.NoError(t, err)
	transferDestination, err := address.Validate("ndantg2f2iz5xs8e3e4ipbk2uef5f98hubvuegn9499udhgm")
	require.NoError(t, err)

	// bmRhbnRnMmYyaXo1eHM4ZTNlNGlwYmsydWVmNWY5OGh1YnZ1ZWduOTQ5OXVkaGdtAAWCIQf8/70AApoPXz8aEm5kYWZtNXA3d2tweXVidms5Y3JnNHlnNmJkbWczcWF2aGhodTZ3azRpZHhyemJzMw==
	expect := []byte{0x6e, 0x64, 0x61, 0x6e, 0x74, 0x67, 0x32, 0x66, 0x32, 0x69, 0x7a, 0x35, 0x78, 0x73, 0x38, 0x65, 0x33, 0x65, 0x34, 0x69, 0x70, 0x62, 0x6b, 0x32, 0x75, 0x65, 0x66, 0x35, 0x66, 0x39, 0x38, 0x68, 0x75, 0x62, 0x76, 0x75, 0x65, 0x67, 0x6e, 0x39, 0x34, 0x39, 0x39, 0x75, 0x64, 0x68, 0x67, 0x6d, 0x00, 0x05, 0x82, 0x21, 0x07, 0xfc, 0xff, 0xbd, 0x00, 0x02, 0x9a, 0x0f, 0x5f, 0x3f, 0x1a, 0x12, 0x6e, 0x64, 0x61, 0x66, 0x6d, 0x35, 0x70, 0x37, 0x77, 0x6b, 0x70, 0x79, 0x75, 0x62, 0x76, 0x6b, 0x39, 0x63, 0x72, 0x67, 0x34, 0x79, 0x67, 0x36, 0x62, 0x64, 0x6d, 0x67, 0x33, 0x71, 0x61, 0x76, 0x68, 0x68, 0x68, 0x75, 0x36, 0x77, 0x6b, 0x34, 0x69, 0x64, 0x78, 0x72, 0x7a, 0x62, 0x73, 0x33}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *Transfer
	}{
		{
			"no signatures",
			NewTransfer(
				transferSource,
				transferDestination,
				1550453263105981,
				732340766579218,
			),
		},
		{
			"with signature",
			NewTransfer(
				transferSource,
				transferDestination,
				1550453263105981,
				732340766579218,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestChangeValidation_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	changevalidationTarget, err := address.Validate("ndaigsueiaggxy4xw69cqfma6fwfijqdmnzp75pjhu8pak4w")
	require.NoError(t, err)
	changevalidationNewKeys, err := signature.RawPublicKey(signature.Ed25519, []byte{0xc1, 0x9f, 0x05, 0xcb, 0xe1, 0xa3, 0x35, 0x34, 0x8c, 0xb5, 0x51, 0xd5, 0xe0, 0x46, 0x74, 0x6c, 0x05, 0xbe, 0x13, 0x8f, 0x3a, 0x81, 0x93, 0x08, 0xa8, 0x22, 0x3f, 0x6d, 0x53, 0xce, 0x99, 0xb8}, nil)
	require.NoError(t, err)

	// bnB1YmE4amFkdGJiZWRhMzhicW02Z3R2a3Blbnl4aTdtMmNncXR5YW1yc3Z0NjdpZGUyaXhhdGQ4NWt2MzRuNXNtNW50d2MyZDluNgAQTcdmlXTFbmRhaWdzdWVpYWdneHk0eHc2OWNxZm1hNmZ3ZmlqcWRtbnpwNzVwamh1OHBhazR3enhlVUpTYUVpclFjeGt6dw==
	expect := []byte{0x6e, 0x70, 0x75, 0x62, 0x61, 0x38, 0x6a, 0x61, 0x64, 0x74, 0x62, 0x62, 0x65, 0x64, 0x61, 0x33, 0x38, 0x62, 0x71, 0x6d, 0x36, 0x67, 0x74, 0x76, 0x6b, 0x70, 0x65, 0x6e, 0x79, 0x78, 0x69, 0x37, 0x6d, 0x32, 0x63, 0x67, 0x71, 0x74, 0x79, 0x61, 0x6d, 0x72, 0x73, 0x76, 0x74, 0x36, 0x37, 0x69, 0x64, 0x65, 0x32, 0x69, 0x78, 0x61, 0x74, 0x64, 0x38, 0x35, 0x6b, 0x76, 0x33, 0x34, 0x6e, 0x35, 0x73, 0x6d, 0x35, 0x6e, 0x74, 0x77, 0x63, 0x32, 0x64, 0x39, 0x6e, 0x36, 0x00, 0x10, 0x4d, 0xc7, 0x66, 0x95, 0x74, 0xc5, 0x6e, 0x64, 0x61, 0x69, 0x67, 0x73, 0x75, 0x65, 0x69, 0x61, 0x67, 0x67, 0x78, 0x79, 0x34, 0x78, 0x77, 0x36, 0x39, 0x63, 0x71, 0x66, 0x6d, 0x61, 0x36, 0x66, 0x77, 0x66, 0x69, 0x6a, 0x71, 0x64, 0x6d, 0x6e, 0x7a, 0x70, 0x37, 0x35, 0x70, 0x6a, 0x68, 0x75, 0x38, 0x70, 0x61, 0x6b, 0x34, 0x77, 0x7a, 0x78, 0x65, 0x55, 0x4a, 0x53, 0x61, 0x45, 0x69, 0x72, 0x51, 0x63, 0x78, 0x6b, 0x7a, 0x77}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *ChangeValidation
	}{
		{
			"no signatures",
			NewChangeValidation(
				changevalidationTarget,
				[]signature.PublicKey{*changevalidationNewKeys},
				// ValidationScript as b64: zxeUJSaEirQcxkzw
				[]byte{0xcf, 0x17, 0x94, 0x25, 0x26, 0x84, 0x8a, 0xb4, 0x1c, 0xc6, 0x4c, 0xf0},
				4589118442271941,
			),
		},
		{
			"with signature",
			NewChangeValidation(
				changevalidationTarget,
				[]signature.PublicKey{*changevalidationNewKeys},
				// ValidationScript as b64: zxeUJSaEirQcxkzw
				[]byte{0xcf, 0x17, 0x94, 0x25, 0x26, 0x84, 0x8a, 0xb4, 0x1c, 0xc6, 0x4c, 0xf0},
				4589118442271941,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestReleaseFromEndowment_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	releasefromendowmentDestination, err := address.Validate("ndarrpbk2278zzdkpmds4vu52buszzvkav78cnvh88qqg9mf")
	require.NoError(t, err)

	// bmRhcnJwYmsyMjc4enpka3BtZHM0dnU1MmJ1c3p6dmthdjc4Y252aDg4cXFnOW1mAAcEu3tNfcMAGoWBScbkgw==
	expect := []byte{0x6e, 0x64, 0x61, 0x72, 0x72, 0x70, 0x62, 0x6b, 0x32, 0x32, 0x37, 0x38, 0x7a, 0x7a, 0x64, 0x6b, 0x70, 0x6d, 0x64, 0x73, 0x34, 0x76, 0x75, 0x35, 0x32, 0x62, 0x75, 0x73, 0x7a, 0x7a, 0x76, 0x6b, 0x61, 0x76, 0x37, 0x38, 0x63, 0x6e, 0x76, 0x68, 0x38, 0x38, 0x71, 0x71, 0x67, 0x39, 0x6d, 0x66, 0x00, 0x07, 0x04, 0xbb, 0x7b, 0x4d, 0x7d, 0xc3, 0x00, 0x1a, 0x85, 0x81, 0x49, 0xc6, 0xe4, 0x83}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *ReleaseFromEndowment
	}{
		{
			"no signatures",
			NewReleaseFromEndowment(
				releasefromendowmentDestination,
				1975528111046083,
				7465139729523843,
			),
		},
		{
			"with signature",
			NewReleaseFromEndowment(
				releasefromendowmentDestination,
				1975528111046083,
				7465139729523843,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestChangeSettlementPeriod_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	changesettlementperiodTarget, err := address.Validate("ndaqczchb8rihvnadyvj88x68ngfx7p3xf2kzhznaeh34qtr")
	require.NoError(t, err)

	// MTFtMjVkdDE0aDM1bTQ5czE2NzMyMHVzAAbLOXkWAGFuZGFxY3pjaGI4cmlodm5hZHl2ajg4eDY4bmdmeDdwM3hmMmt6aHpuYWVoMzRxdHI=
	expect := []byte{0x31, 0x31, 0x6d, 0x32, 0x35, 0x64, 0x74, 0x31, 0x34, 0x68, 0x33, 0x35, 0x6d, 0x34, 0x39, 0x73, 0x31, 0x36, 0x37, 0x33, 0x32, 0x30, 0x75, 0x73, 0x00, 0x06, 0xcb, 0x39, 0x79, 0x16, 0x00, 0x61, 0x6e, 0x64, 0x61, 0x71, 0x63, 0x7a, 0x63, 0x68, 0x62, 0x38, 0x72, 0x69, 0x68, 0x76, 0x6e, 0x61, 0x64, 0x79, 0x76, 0x6a, 0x38, 0x38, 0x78, 0x36, 0x38, 0x6e, 0x67, 0x66, 0x78, 0x37, 0x70, 0x33, 0x78, 0x66, 0x32, 0x6b, 0x7a, 0x68, 0x7a, 0x6e, 0x61, 0x65, 0x68, 0x33, 0x34, 0x71, 0x74, 0x72}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *ChangeSettlementPeriod
	}{
		{
			"no signatures",
			NewChangeSettlementPeriod(
				changesettlementperiodTarget,
				30724549167320,
				1912297565323361,
			),
		},
		{
			"with signature",
			NewChangeSettlementPeriod(
				changesettlementperiodTarget,
				30724549167320,
				1912297565323361,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestDelegate_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	delegateTarget, err := address.Validate("ndahxqn3gbjhhkw6bkxuprhyescywp96vagrm498vf5xpa5b")
	require.NoError(t, err)
	delegateNode, err := address.Validate("ndaaud5dgwnfzpmdpu8ntdkjs4jucye7knbng26vqwfabhzd")
	require.NoError(t, err)

	// bmRhYXVkNWRnd25menBtZHB1OG50ZGtqczRqdWN5ZTdrbmJuZzI2dnF3ZmFiaHpkAAjSrCLE0zRuZGFoeHFuM2diamhoa3c2Ymt4dXByaHllc2N5d3A5NnZhZ3JtNDk4dmY1eHBhNWI=
	expect := []byte{0x6e, 0x64, 0x61, 0x61, 0x75, 0x64, 0x35, 0x64, 0x67, 0x77, 0x6e, 0x66, 0x7a, 0x70, 0x6d, 0x64, 0x70, 0x75, 0x38, 0x6e, 0x74, 0x64, 0x6b, 0x6a, 0x73, 0x34, 0x6a, 0x75, 0x63, 0x79, 0x65, 0x37, 0x6b, 0x6e, 0x62, 0x6e, 0x67, 0x32, 0x36, 0x76, 0x71, 0x77, 0x66, 0x61, 0x62, 0x68, 0x7a, 0x64, 0x00, 0x08, 0xd2, 0xac, 0x22, 0xc4, 0xd3, 0x34, 0x6e, 0x64, 0x61, 0x68, 0x78, 0x71, 0x6e, 0x33, 0x67, 0x62, 0x6a, 0x68, 0x68, 0x6b, 0x77, 0x36, 0x62, 0x6b, 0x78, 0x75, 0x70, 0x72, 0x68, 0x79, 0x65, 0x73, 0x63, 0x79, 0x77, 0x70, 0x39, 0x36, 0x76, 0x61, 0x67, 0x72, 0x6d, 0x34, 0x39, 0x38, 0x76, 0x66, 0x35, 0x78, 0x70, 0x61, 0x35, 0x62}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *Delegate
	}{
		{
			"no signatures",
			NewDelegate(
				delegateTarget,
				delegateNode,
				2483436573217588,
			),
		},
		{
			"with signature",
			NewDelegate(
				delegateTarget,
				delegateNode,
				2483436573217588,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestCreditEAI_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	crediteaiNode, err := address.Validate("ndarg7fz2rhzr8c6tpgazecjmtg96wff9si6i9xz47kgh8vp")
	require.NoError(t, err)

	// bmRhcmc3Znoycmh6cjhjNnRwZ2F6ZWNqbXRnOTZ3ZmY5c2k2aTl4ejQ3a2doOHZwAAl/SDkpCAM=
	expect := []byte{0x6e, 0x64, 0x61, 0x72, 0x67, 0x37, 0x66, 0x7a, 0x32, 0x72, 0x68, 0x7a, 0x72, 0x38, 0x63, 0x36, 0x74, 0x70, 0x67, 0x61, 0x7a, 0x65, 0x63, 0x6a, 0x6d, 0x74, 0x67, 0x39, 0x36, 0x77, 0x66, 0x66, 0x39, 0x73, 0x69, 0x36, 0x69, 0x39, 0x78, 0x7a, 0x34, 0x37, 0x6b, 0x67, 0x68, 0x38, 0x76, 0x70, 0x00, 0x09, 0x7f, 0x48, 0x39, 0x29, 0x08, 0x03}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *CreditEAI
	}{
		{
			"no signatures",
			NewCreditEAI(
				crediteaiNode,
				2673222963759107,
			),
		},
		{
			"with signature",
			NewCreditEAI(
				crediteaiNode,
				2673222963759107,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestLock_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	lockTarget, err := address.Validate("ndaryffgwbces3qxcxcne9i5bhw2e5ca2h9cvzyfmtrrxu7q")
	require.NoError(t, err)

	// M3kyMGR0Nmg0OG0yNzk0NDl1cwASfMbzh18TbmRhcnlmZmd3YmNlczNxeGN4Y25lOWk1Ymh3MmU1Y2EyaDljdnp5Zm10cnJ4dTdx
	expect := []byte{0x33, 0x79, 0x32, 0x30, 0x64, 0x74, 0x36, 0x68, 0x34, 0x38, 0x6d, 0x32, 0x37, 0x39, 0x34, 0x34, 0x39, 0x75, 0x73, 0x00, 0x12, 0x7c, 0xc6, 0xf3, 0x87, 0x5f, 0x13, 0x6e, 0x64, 0x61, 0x72, 0x79, 0x66, 0x66, 0x67, 0x77, 0x62, 0x63, 0x65, 0x73, 0x33, 0x71, 0x78, 0x63, 0x78, 0x63, 0x6e, 0x65, 0x39, 0x69, 0x35, 0x62, 0x68, 0x77, 0x32, 0x65, 0x35, 0x63, 0x61, 0x32, 0x68, 0x39, 0x63, 0x76, 0x7a, 0x79, 0x66, 0x6d, 0x74, 0x72, 0x72, 0x78, 0x75, 0x37, 0x71}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *Lock
	}{
		{
			"no signatures",
			NewLock(
				lockTarget,
				96360480279449,
				5203743511895827,
			),
		},
		{
			"with signature",
			NewLock(
				lockTarget,
				96360480279449,
				5203743511895827,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestNotify_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	notifyTarget, err := address.Validate("ndaadp7vsusgpu5ndzbpm3fpje2snv5x78cuz7a57eptqqzv")
	require.NoError(t, err)

	// AAVar5XpSWhuZGFhZHA3dnN1c2dwdTVuZHpicG0zZnBqZTJzbnY1eDc4Y3V6N2E1N2VwdHFxenY=
	expect := []byte{0x00, 0x05, 0x5a, 0xaf, 0x95, 0xe9, 0x49, 0x68, 0x6e, 0x64, 0x61, 0x61, 0x64, 0x70, 0x37, 0x76, 0x73, 0x75, 0x73, 0x67, 0x70, 0x75, 0x35, 0x6e, 0x64, 0x7a, 0x62, 0x70, 0x6d, 0x33, 0x66, 0x70, 0x6a, 0x65, 0x32, 0x73, 0x6e, 0x76, 0x35, 0x78, 0x37, 0x38, 0x63, 0x75, 0x7a, 0x37, 0x61, 0x35, 0x37, 0x65, 0x70, 0x74, 0x71, 0x71, 0x7a, 0x76}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *Notify
	}{
		{
			"no signatures",
			NewNotify(
				notifyTarget,
				1507085064423784,
			),
		},
		{
			"with signature",
			NewNotify(
				notifyTarget,
				1507085064423784,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestSetRewardsDestination_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	setrewardsdestinationSource, err := address.Validate("ndaqwt9scccvthjtztzmafe64yam9pxzkmysj9s7z5ipb3fg")
	require.NoError(t, err)
	setrewardsdestinationDestination, err := address.Validate("ndak7vmic7pbkv5rjywbnegy6ergjwbnyju9p6g7znptk89k")
	require.NoError(t, err)

	// bmRhazd2bWljN3Bia3Y1cmp5d2JuZWd5NmVyZ2p3Ym55anU5cDZnN3pucHRrODlrAAz/orpRehFuZGFxd3Q5c2NjY3Z0aGp0enR6bWFmZTY0eWFtOXB4emtteXNqOXM3ejVpcGIzZmc=
	expect := []byte{0x6e, 0x64, 0x61, 0x6b, 0x37, 0x76, 0x6d, 0x69, 0x63, 0x37, 0x70, 0x62, 0x6b, 0x76, 0x35, 0x72, 0x6a, 0x79, 0x77, 0x62, 0x6e, 0x65, 0x67, 0x79, 0x36, 0x65, 0x72, 0x67, 0x6a, 0x77, 0x62, 0x6e, 0x79, 0x6a, 0x75, 0x39, 0x70, 0x36, 0x67, 0x37, 0x7a, 0x6e, 0x70, 0x74, 0x6b, 0x38, 0x39, 0x6b, 0x00, 0x0c, 0xff, 0xa2, 0xba, 0x51, 0x7a, 0x11, 0x6e, 0x64, 0x61, 0x71, 0x77, 0x74, 0x39, 0x73, 0x63, 0x63, 0x63, 0x76, 0x74, 0x68, 0x6a, 0x74, 0x7a, 0x74, 0x7a, 0x6d, 0x61, 0x66, 0x65, 0x36, 0x34, 0x79, 0x61, 0x6d, 0x39, 0x70, 0x78, 0x7a, 0x6b, 0x6d, 0x79, 0x73, 0x6a, 0x39, 0x73, 0x37, 0x7a, 0x35, 0x69, 0x70, 0x62, 0x33, 0x66, 0x67}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *SetRewardsDestination
	}{
		{
			"no signatures",
			NewSetRewardsDestination(
				setrewardsdestinationSource,
				setrewardsdestinationDestination,
				3658774096214545,
			),
		},
		{
			"with signature",
			NewSetRewardsDestination(
				setrewardsdestinationSource,
				setrewardsdestinationDestination,
				3658774096214545,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestClaimAccount_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	claimaccountTarget, err := address.Validate("ndaptj5nw6uf3qyfu8ff8g8t9rbspcxjzs9mhxjtpdxkqcqp")
	require.NoError(t, err)
	claimaccountOwnership, err := signature.RawPublicKey(signature.Ed25519, []byte{0x16, 0x77, 0x53, 0x3c, 0xf0, 0x1d, 0x79, 0x69, 0x25, 0x48, 0x13, 0xe4, 0x30, 0x23, 0x6f, 0x1a, 0xef, 0x6f, 0xcd, 0x15, 0x45, 0x17, 0x06, 0x11, 0x73, 0x1e, 0xdf, 0x8f, 0xcf, 0x2b, 0x77, 0x71}, nil)
	require.NoError(t, err)
	claimaccountValidationKeys, err := signature.RawPublicKey(signature.Ed25519, []byte{0x21, 0x4c, 0xbf, 0x93, 0x4b, 0x19, 0x37, 0x6a, 0xd9, 0x05, 0xc3, 0xb8, 0x16, 0xce, 0x67, 0x7d, 0x3e, 0x24, 0xd1, 0xe3, 0xbb, 0x78, 0xef, 0xcb, 0x45, 0x22, 0x85, 0xbd, 0x7b, 0x72, 0xbe, 0x0e}, nil)
	require.NoError(t, err)

	// bnB1YmE4amFkdGJiZWFtaHF3MzY4YXF6dTRqZmphajhpbmJkcDZwcTg1OHBjeGN0cWJzdHFucnA5ZDhyZnA1emRqbjNmYzdmdmd4cQAZX/GDxHNZbmRhcHRqNW53NnVmM3F5ZnU4ZmY4Zzh0OXJic3BjeGp6czltaHhqdHBkeGtxY3FwbnB1YmE4amFkdGJiZWFzdzNyNnZqbm52cTR5M2F6YjVzZnlxbjc4djZqZ3Q2cTd6dDU4bWl3dGltcm01cWs5YTd4Z3d6ZnRqbmt5ZmZOVUtqR05iK21Gc0dvV1k=
	expect := []byte{0x6e, 0x70, 0x75, 0x62, 0x61, 0x38, 0x6a, 0x61, 0x64, 0x74, 0x62, 0x62, 0x65, 0x61, 0x6d, 0x68, 0x71, 0x77, 0x33, 0x36, 0x38, 0x61, 0x71, 0x7a, 0x75, 0x34, 0x6a, 0x66, 0x6a, 0x61, 0x6a, 0x38, 0x69, 0x6e, 0x62, 0x64, 0x70, 0x36, 0x70, 0x71, 0x38, 0x35, 0x38, 0x70, 0x63, 0x78, 0x63, 0x74, 0x71, 0x62, 0x73, 0x74, 0x71, 0x6e, 0x72, 0x70, 0x39, 0x64, 0x38, 0x72, 0x66, 0x70, 0x35, 0x7a, 0x64, 0x6a, 0x6e, 0x33, 0x66, 0x63, 0x37, 0x66, 0x76, 0x67, 0x78, 0x71, 0x00, 0x19, 0x5f, 0xf1, 0x83, 0xc4, 0x73, 0x59, 0x6e, 0x64, 0x61, 0x70, 0x74, 0x6a, 0x35, 0x6e, 0x77, 0x36, 0x75, 0x66, 0x33, 0x71, 0x79, 0x66, 0x75, 0x38, 0x66, 0x66, 0x38, 0x67, 0x38, 0x74, 0x39, 0x72, 0x62, 0x73, 0x70, 0x63, 0x78, 0x6a, 0x7a, 0x73, 0x39, 0x6d, 0x68, 0x78, 0x6a, 0x74, 0x70, 0x64, 0x78, 0x6b, 0x71, 0x63, 0x71, 0x70, 0x6e, 0x70, 0x75, 0x62, 0x61, 0x38, 0x6a, 0x61, 0x64, 0x74, 0x62, 0x62, 0x65, 0x61, 0x73, 0x77, 0x33, 0x72, 0x36, 0x76, 0x6a, 0x6e, 0x6e, 0x76, 0x71, 0x34, 0x79, 0x33, 0x61, 0x7a, 0x62, 0x35, 0x73, 0x66, 0x79, 0x71, 0x6e, 0x37, 0x38, 0x76, 0x36, 0x6a, 0x67, 0x74, 0x36, 0x71, 0x37, 0x7a, 0x74, 0x35, 0x38, 0x6d, 0x69, 0x77, 0x74, 0x69, 0x6d, 0x72, 0x6d, 0x35, 0x71, 0x6b, 0x39, 0x61, 0x37, 0x78, 0x67, 0x77, 0x7a, 0x66, 0x74, 0x6a, 0x6e, 0x6b, 0x79, 0x66, 0x66, 0x4e, 0x55, 0x4b, 0x6a, 0x47, 0x4e, 0x62, 0x2b, 0x6d, 0x46, 0x73, 0x47, 0x6f, 0x57, 0x59}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *ClaimAccount
	}{
		{
			"no signatures",
			NewClaimAccount(
				claimaccountTarget,
				*claimaccountOwnership,
				[]signature.PublicKey{*claimaccountValidationKeys},
				// ValidationScript as b64: fNUKjGNb+mFsGoWY
				[]byte{0x7c, 0xd5, 0x0a, 0x8c, 0x63, 0x5b, 0xfa, 0x61, 0x6c, 0x1a, 0x85, 0x98},
				7142365320213337,
			),
		},
		{
			"with signature",
			NewClaimAccount(
				claimaccountTarget,
				*claimaccountOwnership,
				[]signature.PublicKey{*claimaccountValidationKeys},
				// ValidationScript as b64: fNUKjGNb+mFsGoWY
				[]byte{0x7c, 0xd5, 0x0a, 0x8c, 0x63, 0x5b, 0xfa, 0x61, 0x6c, 0x1a, 0x85, 0x98},
				7142365320213337,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestStake_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	stakeTarget, err := address.Validate("ndacnipyesepxdfk8r43cdm6vj5r5mf85gs7ks89w9vp5dqb")
	require.NoError(t, err)
	stakeNode, err := address.Validate("ndaqst4h5iywbg7zrj9wvd2w8b4dpiwxyq4r57v892fag37a")
	require.NoError(t, err)

	// bmRhcXN0NGg1aXl3Ymc3enJqOXd2ZDJ3OGI0ZHBpd3h5cTRyNTd2ODkyZmFnMzdhABYhWHyzrlBuZGFjbmlweWVzZXB4ZGZrOHI0M2NkbTZ2ajVyNW1mODVnczdrczg5dzl2cDVkcWI=
	expect := []byte{0x6e, 0x64, 0x61, 0x71, 0x73, 0x74, 0x34, 0x68, 0x35, 0x69, 0x79, 0x77, 0x62, 0x67, 0x37, 0x7a, 0x72, 0x6a, 0x39, 0x77, 0x76, 0x64, 0x32, 0x77, 0x38, 0x62, 0x34, 0x64, 0x70, 0x69, 0x77, 0x78, 0x79, 0x71, 0x34, 0x72, 0x35, 0x37, 0x76, 0x38, 0x39, 0x32, 0x66, 0x61, 0x67, 0x33, 0x37, 0x61, 0x00, 0x16, 0x21, 0x58, 0x7c, 0xb3, 0xae, 0x50, 0x6e, 0x64, 0x61, 0x63, 0x6e, 0x69, 0x70, 0x79, 0x65, 0x73, 0x65, 0x70, 0x78, 0x64, 0x66, 0x6b, 0x38, 0x72, 0x34, 0x33, 0x63, 0x64, 0x6d, 0x36, 0x76, 0x6a, 0x35, 0x72, 0x35, 0x6d, 0x66, 0x38, 0x35, 0x67, 0x73, 0x37, 0x6b, 0x73, 0x38, 0x39, 0x77, 0x39, 0x76, 0x70, 0x35, 0x64, 0x71, 0x62}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *Stake
	}{
		{
			"no signatures",
			NewStake(
				stakeTarget,
				stakeNode,
				6229113420623440,
			),
		},
		{
			"with signature",
			NewStake(
				stakeTarget,
				stakeNode,
				6229113420623440,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestRegisterNode_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	registernodeNode, err := address.Validate("ndaefznjujwi2nsk687ukxzdk5ijt7bh6wvd86zdxaas8cud")
	require.NoError(t, err)

	// VDBZR0JiMEI5VU5kQ0g4dm5kYWVmem5qdWp3aTJuc2s2ODd1a3h6ZGs1aWp0N2JoNnd2ZDg2emR4YWFzOGN1ZHN0cmluZzogdGhjdCB1YXggeGtxZiBhZnBsc2ogYmN4b2VmZnIgABQD3vY6XHI=
	expect := []byte{0x54, 0x30, 0x59, 0x47, 0x42, 0x62, 0x30, 0x42, 0x39, 0x55, 0x4e, 0x64, 0x43, 0x48, 0x38, 0x76, 0x6e, 0x64, 0x61, 0x65, 0x66, 0x7a, 0x6e, 0x6a, 0x75, 0x6a, 0x77, 0x69, 0x32, 0x6e, 0x73, 0x6b, 0x36, 0x38, 0x37, 0x75, 0x6b, 0x78, 0x7a, 0x64, 0x6b, 0x35, 0x69, 0x6a, 0x74, 0x37, 0x62, 0x68, 0x36, 0x77, 0x76, 0x64, 0x38, 0x36, 0x7a, 0x64, 0x78, 0x61, 0x61, 0x73, 0x38, 0x63, 0x75, 0x64, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x3a, 0x20, 0x74, 0x68, 0x63, 0x74, 0x20, 0x75, 0x61, 0x78, 0x20, 0x78, 0x6b, 0x71, 0x66, 0x20, 0x61, 0x66, 0x70, 0x6c, 0x73, 0x6a, 0x20, 0x62, 0x63, 0x78, 0x6f, 0x65, 0x66, 0x66, 0x72, 0x20, 0x00, 0x14, 0x03, 0xde, 0xf6, 0x3a, 0x5c, 0x72}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *RegisterNode
	}{
		{
			"no signatures",
			NewRegisterNode(
				registernodeNode,
				// DistributionScript as b64: T0YGBb0B9UNdCH8v
				[]byte{0x4f, 0x46, 0x06, 0x05, 0xbd, 0x01, 0xf5, 0x43, 0x5d, 0x08, 0x7f, 0x2f},
				"string: thct uax xkqf afplsj bcxoeffr ",
				5633755682856050,
			),
		},
		{
			"with signature",
			NewRegisterNode(
				registernodeNode,
				// DistributionScript as b64: T0YGBb0B9UNdCH8v
				[]byte{0x4f, 0x46, 0x06, 0x05, 0xbd, 0x01, 0xf5, 0x43, 0x5d, 0x08, 0x7f, 0x2f},
				"string: thct uax xkqf afplsj bcxoeffr ",
				5633755682856050,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestNominateNodeReward_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// ABs65BlHbVYAEQF/vImALQ==
	expect := []byte{0x00, 0x1b, 0x3a, 0xe4, 0x19, 0x47, 0x6d, 0x56, 0x00, 0x11, 0x01, 0x7f, 0xbc, 0x89, 0x80, 0x2d}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *NominateNodeReward
	}{
		{
			"no signatures",
			NewNominateNodeReward(
				7664575722253654,
				4786722739683373,
			),
		},
		{
			"with signature",
			NewNominateNodeReward(
				7664575722253654,
				4786722739683373,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestClaimNodeReward_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	claimnoderewardNode, err := address.Validate("ndahr4crb57v7i8uihd92ytmx7gap36mj7jhasu3ypdqw8uz")
	require.NoError(t, err)

	// bmRhaHI0Y3JiNTd2N2k4dWloZDkyeXRteDdnYXAzNm1qN2poYXN1M3lwZHF3OHV6AA+kWRaYdWc=
	expect := []byte{0x6e, 0x64, 0x61, 0x68, 0x72, 0x34, 0x63, 0x72, 0x62, 0x35, 0x37, 0x76, 0x37, 0x69, 0x38, 0x75, 0x69, 0x68, 0x64, 0x39, 0x32, 0x79, 0x74, 0x6d, 0x78, 0x37, 0x67, 0x61, 0x70, 0x33, 0x36, 0x6d, 0x6a, 0x37, 0x6a, 0x68, 0x61, 0x73, 0x75, 0x33, 0x79, 0x70, 0x64, 0x71, 0x77, 0x38, 0x75, 0x7a, 0x00, 0x0f, 0xa4, 0x59, 0x16, 0x98, 0x75, 0x67}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *ClaimNodeReward
	}{
		{
			"no signatures",
			NewClaimNodeReward(
				claimnoderewardNode,
				4402827188794727,
			),
		},
		{
			"with signature",
			NewClaimNodeReward(
				claimnoderewardNode,
				4402827188794727,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestTransferAndLock_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	transferandlockSource, err := address.Validate("ndacrr6m2p8amtcb2jdys5mi95e7bhtt6v8tdacgct6b77i5")
	require.NoError(t, err)
	transferandlockDestination, err := address.Validate("ndah28g6nggny49cq4g3cemef52f74pyrnjt7vvvn8vh4gud")
	require.NoError(t, err)

	// bmRhaDI4ZzZuZ2dueTQ5Y3E0ZzNjZW1lZjUyZjc0cHlybmp0N3Z2dm44dmg0Z3VkMXk5bTE1ZHQxMGg5bTUxczU2OTcyMXVzABWjFXJMfnYAGnH1Ile/EG5kYWNycjZtMnA4YW10Y2IyamR5czVtaTk1ZTdiaHR0NnY4dGRhY2djdDZiNzdpNQ==
	expect := []byte{0x6e, 0x64, 0x61, 0x68, 0x32, 0x38, 0x67, 0x36, 0x6e, 0x67, 0x67, 0x6e, 0x79, 0x34, 0x39, 0x63, 0x71, 0x34, 0x67, 0x33, 0x63, 0x65, 0x6d, 0x65, 0x66, 0x35, 0x32, 0x66, 0x37, 0x34, 0x70, 0x79, 0x72, 0x6e, 0x6a, 0x74, 0x37, 0x76, 0x76, 0x76, 0x6e, 0x38, 0x76, 0x68, 0x34, 0x67, 0x75, 0x64, 0x31, 0x79, 0x39, 0x6d, 0x31, 0x35, 0x64, 0x74, 0x31, 0x30, 0x68, 0x39, 0x6d, 0x35, 0x31, 0x73, 0x35, 0x36, 0x39, 0x37, 0x32, 0x31, 0x75, 0x73, 0x00, 0x15, 0xa3, 0x15, 0x72, 0x4c, 0x7e, 0x76, 0x00, 0x1a, 0x71, 0xf5, 0x22, 0x57, 0xbf, 0x10, 0x6e, 0x64, 0x61, 0x63, 0x72, 0x72, 0x36, 0x6d, 0x32, 0x70, 0x38, 0x61, 0x6d, 0x74, 0x63, 0x62, 0x32, 0x6a, 0x64, 0x79, 0x73, 0x35, 0x6d, 0x69, 0x39, 0x35, 0x65, 0x37, 0x62, 0x68, 0x74, 0x74, 0x36, 0x76, 0x38, 0x74, 0x64, 0x61, 0x63, 0x67, 0x63, 0x74, 0x36, 0x62, 0x37, 0x37, 0x69, 0x35}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *TransferAndLock
	}{
		{
			"no signatures",
			NewTransferAndLock(
				transferandlockSource,
				transferandlockDestination,
				6090287018180214,
				56196591569721,
				7443647051579152,
			),
		},
		{
			"with signature",
			NewTransferAndLock(
				transferandlockSource,
				transferandlockDestination,
				6090287018180214,
				56196591569721,
				7443647051579152,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestCommandValidatorChange_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// ABlkl28mmotUVFYra2QwRFl6RG1ybml5AAucqL6OnWk=
	expect := []byte{0x00, 0x19, 0x64, 0x97, 0x6f, 0x26, 0x9a, 0x8b, 0x54, 0x54, 0x56, 0x2b, 0x6b, 0x64, 0x30, 0x44, 0x59, 0x7a, 0x44, 0x6d, 0x72, 0x6e, 0x69, 0x79, 0x00, 0x0b, 0x9c, 0xa8, 0xbe, 0x8e, 0x9d, 0x69}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *CommandValidatorChange
	}{
		{
			"no signatures",
			NewCommandValidatorChange(
				// PublicKey as b64: TTV+kd0DYzDmrniy
				[]byte{0x4d, 0x35, 0x7e, 0x91, 0xdd, 0x03, 0x63, 0x30, 0xe6, 0xae, 0x78, 0xb2},
				7147475985406603,
				3268473309273449,
			),
		},
		{
			"with signature",
			NewCommandValidatorChange(
				// PublicKey as b64: TTV+kd0DYzDmrniy
				[]byte{0x4d, 0x35, 0x7e, 0x91, 0xdd, 0x03, 0x63, 0x30, 0xe6, 0xae, 0x78, 0xb2},
				7147475985406603,
				3268473309273449,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}
func TestSidechainTx_SignableBytes(t *testing.T) {
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	sidechaintxSource, err := address.Validate("ndad32zv2cugaydh2gtycxgvny8esp5vk3i67dawrqpck7a2")
	require.NoError(t, err)
	sidechaintxSidechainSignatures, err := signature.RawSignature(signature.Ed25519, []byte{0x05, 0x8a, 0x14, 0x9f, 0xb4, 0x5e, 0xe6, 0xc3, 0x2f, 0x97, 0xa4, 0x45, 0x65, 0x51, 0x4b, 0x39, 0x8f, 0x06, 0x15, 0xb0, 0xde, 0xf5, 0xc5, 0x14, 0xca, 0x15, 0xd5, 0x71, 0x50, 0x2d, 0x23, 0x1a, 0x88, 0x3c, 0x7d, 0x82, 0xed, 0x06, 0x78, 0x28, 0x1e, 0x4b, 0x83, 0x1d, 0x56, 0x68, 0x66, 0x38, 0x01, 0xc7, 0xcf, 0x66, 0xa3, 0xf8, 0xf9, 0x1b, 0x11, 0xe1, 0x29, 0x68, 0x97, 0x65, 0x41, 0x9e})
	require.NoError(t, err)

	// AAJeEXA5N5cAAAAAAAAAskNHcGk5VGkxYW1CSGZBZFJhNGphZHRjYWF5ZmJqaDd3bTV2bmdtNnp3dGN5a3drbWhnaHNuZnBzNTU0NmtmZ2tjemt6Y3dicGVucGlzcmQ3c215c244YmlkM2YyZ2hreXBidmRzYXFoMzd2a2g4aDNkbmk4Y2ttaXU3dXdkaHd1ZTI1eTJjZzVuZGFkMzJ6djJjdWdheWRoMmd0eWN4Z3ZueThlc3A1dmszaTY3ZGF3cnFwY2s3YTI=
	expect := []byte{0x00, 0x02, 0x5e, 0x11, 0x70, 0x39, 0x37, 0x97, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xb2, 0x43, 0x47, 0x70, 0x69, 0x39, 0x54, 0x69, 0x31, 0x61, 0x6d, 0x42, 0x48, 0x66, 0x41, 0x64, 0x52, 0x61, 0x34, 0x6a, 0x61, 0x64, 0x74, 0x63, 0x61, 0x61, 0x79, 0x66, 0x62, 0x6a, 0x68, 0x37, 0x77, 0x6d, 0x35, 0x76, 0x6e, 0x67, 0x6d, 0x36, 0x7a, 0x77, 0x74, 0x63, 0x79, 0x6b, 0x77, 0x6b, 0x6d, 0x68, 0x67, 0x68, 0x73, 0x6e, 0x66, 0x70, 0x73, 0x35, 0x35, 0x34, 0x36, 0x6b, 0x66, 0x67, 0x6b, 0x63, 0x7a, 0x6b, 0x7a, 0x63, 0x77, 0x62, 0x70, 0x65, 0x6e, 0x70, 0x69, 0x73, 0x72, 0x64, 0x37, 0x73, 0x6d, 0x79, 0x73, 0x6e, 0x38, 0x62, 0x69, 0x64, 0x33, 0x66, 0x32, 0x67, 0x68, 0x6b, 0x79, 0x70, 0x62, 0x76, 0x64, 0x73, 0x61, 0x71, 0x68, 0x33, 0x37, 0x76, 0x6b, 0x68, 0x38, 0x68, 0x33, 0x64, 0x6e, 0x69, 0x38, 0x63, 0x6b, 0x6d, 0x69, 0x75, 0x37, 0x75, 0x77, 0x64, 0x68, 0x77, 0x75, 0x65, 0x32, 0x35, 0x79, 0x32, 0x63, 0x67, 0x35, 0x6e, 0x64, 0x61, 0x64, 0x33, 0x32, 0x7a, 0x76, 0x32, 0x63, 0x75, 0x67, 0x61, 0x79, 0x64, 0x68, 0x32, 0x67, 0x74, 0x79, 0x63, 0x78, 0x67, 0x76, 0x6e, 0x79, 0x38, 0x65, 0x73, 0x70, 0x35, 0x76, 0x6b, 0x33, 0x69, 0x36, 0x37, 0x64, 0x61, 0x77, 0x72, 0x71, 0x70, 0x63, 0x6b, 0x37, 0x61, 0x32}
	require.NotEmpty(t, expect, "test not properly set up")

	// note the "want" field for both of these tests is identical
	tests := []struct {
		name string
		tx   *SidechainTx
	}{
		{
			"no signatures",
			NewSidechainTx(
				sidechaintxSource,
				178,
				// SidechainSignableBytes as b64: CGpi9Ti1amBHfAdR
				[]byte{0x08, 0x6a, 0x62, 0xf5, 0x38, 0xb5, 0x6a, 0x60, 0x47, 0x7c, 0x07, 0x51},
				[]signature.Signature{*sidechaintxSidechainSignatures},
				666378943674263,
			),
		},
		{
			"with signature",
			NewSidechainTx(
				sidechaintxSource,
				178,
				// SidechainSignableBytes as b64: CGpi9Ti1amBHfAdR
				[]byte{0x08, 0x6a, 0x62, 0xf5, 0x38, 0xb5, 0x6a, 0x60, 0x47, 0x7c, 0x07, 0x51},
				[]signature.Signature{*sidechaintxSidechainSignatures},
				666378943674263,
				private,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, expect, tt.tx.SignableBytes())
		})
	}
}