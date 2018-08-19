package config

import (
	"encoding/base64"
	"fmt"

	"github.com/oneiro-ndev/signature/pkg/signature"
)

// A Keypair holds a pair of keys
type Keypair struct {
	Public  signature.PublicKey  `toml:"public"`
	Private signature.PrivateKey `toml:"private"`
}

// String satisfies io.Stringer by writing the public key
func (kp Keypair) String() string {
	return fmt.Sprintf("%s...", base64.StdEncoding.EncodeToString(kp.Public.Bytes())[:10])
}
