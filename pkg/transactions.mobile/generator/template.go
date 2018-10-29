package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const templateName = "tx.go"

func dereferenced(s string) string {
	for isPointer(s) {
		s = s[1:]
	}
	return s
}

func isPointer(s string) bool {
	return strings.HasPrefix(s, "*")
}

func isSlice(s string) bool {
	return strings.HasPrefix(s, "[]")
}

func singular(s string) string {
	idxLast := len(s) - 1
	if s[idxLast] == 's' || s[idxLast] == 'S' {
		return s[:idxLast]
	}
	return s
}

func unslice(s string) string {
	idxFirst := 0
	if strings.HasPrefix(s, "[]") {
		idxFirst = 2
	}
	return s[idxFirst:]
}

func zero(s string) string {
	return fmt.Sprintf("*new(%s)", s)
}

// ParseTemplate parses the template definition
func ParseTemplate() (*template.Template, error) {
	path := os.ExpandEnv(TemplatePath)

	return template.New(templateName).Funcs(template.FuncMap{
		"Dereferenced": dereferenced,
		"IsPointer":    isPointer,
		"IsSlice":      isSlice,
		"Lower":        strings.ToLower,
		"Singular":     singular,
		"Unslice":      unslice,
		"Zero":         zero,
	}).ParseFiles(path)
}

// ApplyTemplate applies the template to the transaction
func ApplyTemplate(t *template.Template, tx Transaction) error {
	path := filepath.Join(txmobile, strings.ToLower(tx.Name)+".go")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.ExecuteTemplate(f, templateName+".tmpl", tx)
}
