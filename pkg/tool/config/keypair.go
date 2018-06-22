package config

import (
	"encoding/base64"

	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// A Keypair holds a pair of keys
type Keypair struct {
	Public  signature.PublicKey
	Private signature.PrivateKey
}

func (kp Keypair) toToml() (tomlKeypair, error) {
	pubBytes, err := kp.Public.MarshalMsg(nil)
	if err != nil {
		return tomlKeypair{}, errors.Wrap(err, "marshaling public")
	}
	privBytes, err := kp.Private.MarshalMsg(nil)
	if err != nil {
		return tomlKeypair{}, errors.Wrap(err, "marshaling private")
	}

	return tomlKeypair{
		Public:  base64.StdEncoding.EncodeToString(pubBytes),
		Private: base64.StdEncoding.EncodeToString(privBytes),
	}, nil
}

// String satisfies io.Stringer by writing the public key
func (kp Keypair) String() string {
	return base64.StdEncoding.EncodeToString(kp.Public.Bytes())
}

type tomlKeypair struct {
	Public  string `toml:"public"`
	Private string `toml:"private"`
}

func (tkp tomlKeypair) toKeypair() (Keypair, error) {
	pubBytes, err := base64.StdEncoding.DecodeString(tkp.Public)
	if err != nil {
		return Keypair{}, errors.Wrap(err, "unencoding public")
	}
	privBytes, err := base64.StdEncoding.DecodeString(tkp.Private)
	if err != nil {
		return Keypair{}, errors.Wrap(err, "unencoding private")
	}
	pub := signature.PublicKey{}
	_, err = pub.UnmarshalMsg(pubBytes)
	if err != nil {
		return Keypair{}, errors.Wrap(err, "unmarshaling public")
	}
	priv := signature.PrivateKey{}
	_, err = priv.UnmarshalMsg(privBytes)
	if err != nil {
		return Keypair{}, errors.Wrap(err, "unmarshaling private")
	}
	return Keypair{Public: pub, Private: priv}, nil
}
