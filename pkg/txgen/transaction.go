package txgen

// Transaction stores metadata about a transaction
type Transaction struct {
	Name   string
	Fields []Field
}

// HasField is true if the transaction has a field with the specified name
func (t Transaction) HasField(name string) bool {
	for _, f := range t.Fields {
		if f.Name == name {
			return true
		}
	}
	return false
}
