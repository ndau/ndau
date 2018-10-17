package config

import (
	"fmt"

	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

// A Keypair holds a pair of keys
type Keypair struct {
	Public  signature.PublicKey  `toml:"public"`
	Private signature.PrivateKey `toml:"private"`
}

// String satisfies io.Stringer by writing the public key
func (kp Keypair) String() string {
	return fmt.Sprintf("<keypair: pub %s>", kp.Public.String())
}
