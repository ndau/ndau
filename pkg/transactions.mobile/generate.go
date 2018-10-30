package mobile

// To avoid unnecessary imports, go generate has to run a mini-toolchain here:
//   1. generate the wrapper source
//   2. ensure goimports tool is available
//   3. run goimports to clean up imports and format generated code
//
// Step 2 is necessary because we don't want to make any assumptions about what's
// available on a developer machine, and the generated files will fail to compile
// if they include unused imports. It should be fast, though, at least after
// the first time.
//
//go:generate go run $GOPATH/src/github.com/oneiro-ndev/ndau/cmd/tx.mobile/main.go
//go:generate go get golang.org/x/tools/cmd/goimports
//go:generate goimports -w $GOPATH/src/github.com/oneiro-ndev/ndau/pkg/transactions.mobile
