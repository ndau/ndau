package txgen

import (
	"os"
	"strings"
	"text/template"
)

const templateName = "signable_bytes.go"

// ParseTemplate parses the template definition
func ParseTemplate(path string) (*template.Template, error) {
	return template.New(templateName).Funcs(template.FuncMap{
		"Lower": strings.ToLower,
	}).ParseFiles(path)
}

// ApplyTemplate applies the template to the transaction
func ApplyTemplate(t *template.Template, c *Context, output string) error {
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.ExecuteTemplate(f, templateName+".tmpl", *c)
}
