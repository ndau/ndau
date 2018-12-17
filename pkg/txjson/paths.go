package generator

import (
	"os"
	"path/filepath"
)

var (
	rawProject = "$GOPATH/src/github.com/oneiro-ndev/ndau/"
	project    string
	// TransactionPath is the path to the transaction definition file
	TransactionPath string
	txjson          string
	examples        string
	// TemplatePath is the path to the template file
	TemplatePath        string
	jsonLiteralsCmd     string
	jsonLiteralsCmdMain string
)

func init() {
	project = os.ExpandEnv(rawProject)
	TransactionPath = filepath.Join(project, "pkg", "ndau", "transactions.go")
	txjson = filepath.Join(project, "pkg", "txjson")
	examples = filepath.Join(txjson, "examples")
	TemplatePath = filepath.Join(txjson, "json_literals.go.tmpl")
	jsonLiteralsCmd = filepath.Join(project, "cmd", "json_literals")
	jsonLiteralsCmdMain = filepath.Join(jsonLiteralsCmd, "main.go")
}
