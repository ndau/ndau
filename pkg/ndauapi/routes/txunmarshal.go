package routes

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
)

// Given the name of a transaction type and a reader containing the JSON for a transaction
// (usually the request Body from a POST), this constructs a new object containing that
// transactions's data.
func txUnmarshal(txtype string, r io.Reader) (metatx.Transactable, error) {
	// iterate over all the transaction types
	for id, txable := range ndau.TxIDs {
		// find the name of each transaction type
		name := metatx.NameOf(txable)
		// compare to the name we were given
		if strings.EqualFold(name, txtype) {
			// extract the pointer from the table -- it's an interface
			ptr := ndau.TxIDs[id]
			// dereference the pointer to get an object (as an interface)
			txobj := reflect.Indirect(reflect.ValueOf(ptr)).Interface()
			// create a new object that is of the same type (it's also hidden in an interface)
			tx := reflect.New(reflect.TypeOf(txobj)).Interface()
			// decode the stream into the object
			err := json.NewDecoder(r).Decode(tx)
			// cast the result to a Transactable as promised.
			return tx.(metatx.Transactable), err
		}
	}
	return nil, errors.New("no txtype found that matched " + txtype)
}
