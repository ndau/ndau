package ndau

// NdauTransaction is the transaction type that the Ndau chain actually cares about.
//
// Whereas Transaction is the type that tendermint cares about, with nonces
// etc designed to keep consensus running, Transaction handles the Ndau stuff.
type NdauTransaction interface {
	// IsValid returns nil if the NdauTransaction is valid, or an error otherwise.
	IsValid(app *App) error
	// Apply applies this transaction to the supplied application, updating its
	// internal state as required.
	//
	// If anything but nil is returned, the internal state of the input App
	// must be unchanged.
	Apply(app *App) error
	// AsTransaction creates a Transaction from this NdauTransaction.
	AsTransaction() *Transaction
}

// ToNdauTransaction unpacks a Transaction object into an NdauTransaction.
//
// If it does not match any known NdauTransactino
func ToNdauTransaction(tx *Transaction) NdauTransaction {
	switch nt := tx.Tx.(type) {
	case *Transaction_GtValidatorChange:
		return nt.GtValidatorChange
	default:
		return nil
	}
}
