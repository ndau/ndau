package ndau

// run the generators from the "generator" repo

//go:generate go run $GOPATH/src/github.com/oneiro-ndev/generator/cmd/generate
//go:generate go run $GOPATH/src/github.com/oneiro-ndev/generator/cmd/json_literals
//go:generate go run $GOPATH/src/github.com/oneiro-ndev/generator/cmd/maketests
//go:generate tar -cjf $GOPATH/src/github.com/oneiro-ndev/generator/examples.tar.bz2 -C $GOPATH/src/github.com/oneiro-ndev/generator/ examples
//go:generate rm -rf $GOPATH/src/github.com/oneiro-ndev/generator/cmd/json_literals
//go:generate rm -rf $GOPATH/src/github.com/oneiro-ndev/generator/cmd/maketests

//go:generate find $GOPATH/src/github.com/oneiro-ndev/ndau/pkg/ndau/ -name "*_gen*.go" -maxdepth 1 -exec goimports -w {} ;
