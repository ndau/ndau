package txgen

import (
	"os"
	"path/filepath"
)

var (
	rawProject = "$GOPATH/src/github.com/oneiro-ndev/ndau/"
	project    string
	ndauPkg    string
	// TransactionPath is the path to the transaction definition file
	TransactionPath string
	// DefaultOutputPath is the path to the default output file
	DefaultOutputPath string
	// DefaultTemplatePath is the path to the default template
	DefaultTemplatePath string
)

func init() {
	project = os.ExpandEnv(rawProject)
	ndauPkg = filepath.Join(project, "pkg", "ndau")
	TransactionPath = filepath.Join(ndauPkg, "transactions.go")
	DefaultOutputPath = filepath.Join(ndauPkg, "signable_bytes_gen.go")
	DefaultTemplatePath = filepath.Join(project, "pkg", "txgen", "signable_bytes.go.tmpl")
}
