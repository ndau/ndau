package ndau

import "github.com/pkg/errors"

// IsValid implements metatx.Transactable
func (ct *ChangeTransferKey) IsValid(appI interface{}) error {
	return errors.New("not implemented")
}

// Apply implements metatx.Transactable
func (ct *ChangeTransferKey) Apply(appI interface{}) error {
	return errors.New("not implemented")
}
