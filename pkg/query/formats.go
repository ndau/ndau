package query

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// These constants define the format strings which controls the information in the Info field of the relevant queries
const (
	AccountInfoFmt           = "acct exists: %t"
	PrevalidateInfoFmt       = "estimated tx fee: %d napu; estimated sib: %d napu"
	SidechainTxExistsInfoFmt = "sidechain tx paid for and validated: %t"
)
