package generator

import (
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

const templateName = "json_literals.go"

// ParseTemplate parses the template definition
func ParseTemplate() (*template.Template, error) {
	path := os.ExpandEnv(TemplatePath)

	return template.New(templateName).Funcs(template.FuncMap{
		"Lower": strings.ToLower,
	}).ParseFiles(path)
}

// ApplyTemplate applies the template to the transaction
func ApplyTemplate(t *template.Template, c *Context) error {
	err := os.MkdirAll(jsonLiteralsCmd, 0700)
	if err != nil {
		return errors.Wrap(err, "making "+jsonLiteralsCmd)
	}
	f, err := os.Create(jsonLiteralsCmdMain)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.ExecuteTemplate(f, templateName+".tmpl", *c)
}
