package ndau

import (
	"encoding/binary"

	"github.com/oneiro-ndev/signature/pkg/signature"
)

// NewComputeEAI creates a new ComputeEAI transaction
//
// Most users will never need this.
func NewComputeEAI(seed, sequence uint64, key signature.PrivateKey) *ComputeEAI {
	c := &ComputeEAI{Seed: seed, Sequence: sequence}
	c.Signature = key.Sign(c.signableBytes())
	return c
}

func (c *ComputeEAI) signableBytes() []byte {
	bytes := make([]byte, 0, 16)
	binary.BigEndian.PutUint64(bytes, c.Seed)
	binary.BigEndian.PutUint64(bytes, c.Sequence)
	return bytes
}

// Validate implements metatx.Transactable
func (c *ComputeEAI) Validate(appI interface{}) error {
	return nil
}

// Apply implements metatx.Transactable
func (c *ComputeEAI) Apply(appI interface{}) error {
	return nil
}
