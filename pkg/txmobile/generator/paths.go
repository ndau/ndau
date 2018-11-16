package generator

import "os"

var (
	rawProject = "$GOPATH/src/github.com/oneiro-ndev/ndau/"
	project    string
	// TransactionPath is the path to the transaction definition file
	TransactionPath string
	txmobile        string
	generator       string
	// TemplatePath is the path to the transaction template file
	TemplatePath string
)

func init() {
	project = os.ExpandEnv(rawProject)
	TransactionPath = project + "pkg/ndau/transactions.go"
	txmobile = project + "pkg/txmobile/"
	generator = txmobile + "generator/"
	TemplatePath = generator + "tx.go.tmpl"
}
