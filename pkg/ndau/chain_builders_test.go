package ndau

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"reflect"
	"testing"

	"github.com/ndau/chaincode/pkg/vm"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/bitset256"
	math "github.com/ndau/ndaumath/pkg/types"
)

func asm(s string) []byte {
	opcodes := vm.MiniAsm(s)
	b := make([]byte, len(opcodes))
	for i := range opcodes {
		b[i] = byte(opcodes[i])
	}
	return b
}

func TestBuildVMForTxValidation(t *testing.T) {
	type args struct {
		code         string
		acct         backing.AccountData
		tx           metatx.Transactable
		signatureSet *bitset256.Bitset256
	}
	tests := []struct {
		name          string
		args          args
		ts            math.Timestamp
		want          vm.Value
		wantErrCreate bool
		wantErrRun    bool
	}{
		{
			"basic operation",
			args{
				"handler 1 1 push1 9 enddef",
				backing.AccountData{},
				&Transfer{},
				bitset256.New(),
			},
			math.Timestamp(0),
			vm.NewNumber(9),
			false,
			false,
		},
		{
			"test the bitset",
			args{
				"handler 1 1 push1 2 mul enddef",
				backing.AccountData{},
				&Transfer{},
				bitset256.New(3),
			},
			math.Timestamp(0),
			vm.NewNumber(16),
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := initApp(t)

			vm, err := BuildVMForTxValidation(asm(tt.args.code), tt.args.acct, tt.args.tx, tt.args.signatureSet, app)
			if (err != nil) != tt.wantErrCreate {
				t.Errorf("BuildVMForTxValidation() error = %v, wantErrCreate %v", err, tt.wantErrCreate)
				return
			}
			err = vm.Run(nil)
			if (err != nil) != tt.wantErrRun {
				t.Errorf("BuildVMForTxValidation() error = %v, wantErrRun %v", err, tt.wantErrRun)
				return
			}
			got, err := vm.Stack().Pop()
			if err != nil {
				t.Errorf("BuildVMForTxValidation() stack error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildVMForTxValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}
