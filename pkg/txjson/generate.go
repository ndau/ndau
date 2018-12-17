package generator

// use go generate to create json examples for every transaction

//go:generate go run $GOPATH/src/github.com/oneiro-ndev/ndau/cmd/generate_json_literals
//go:generate go run $GOPATH/src/github.com/oneiro-ndev/ndau/cmd/json_literals
//go:generate rm -rf $GOPATH/src/github.com/oneiro-ndev/ndau/cmd/json_literals
//go:generate tar -cjvf $GOPATH/src/github.com/oneiro-ndev/ndau/examples.tar.bz2 -C $GOPATH/src/github.com/oneiro-ndev/ndau/pkg/txjson/examples/ .
