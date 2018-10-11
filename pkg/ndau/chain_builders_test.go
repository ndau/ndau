package ndau

import (
	"reflect"
	"testing"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
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
			app.blockTime = tt.ts

			vm, err := BuildVMForTxValidation(asm(tt.args.code), tt.args.acct, tt.args.tx, tt.args.signatureSet, app)
			if (err != nil) != tt.wantErrCreate {
				t.Errorf("BuildVMForTxValidation() error = %v, wantErrCreate %v", err, tt.wantErrCreate)
				return
			}
			err = vm.Run(false)
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
