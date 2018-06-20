package tool

import (
	"encoding/base64"
	"fmt"
	"io"
)

// EmitIdentityHeader writes an identity header to the specified Writer
func EmitIdentityHeader(out io.Writer) {
	fmt.Fprintf(out, "%-29s %s\n", "NAME", "PUBLIC KEY (base64)")
}

// EmitIdentity writes an identity to the specified Writer
func EmitIdentity(out io.Writer, identity Identity) {
	fmt.Fprintf(
		out, "%-29s %s\n",
		identity.Name,
		base64.StdEncoding.EncodeToString(identity.PublicKey),
	)
}

// EmitIdentities writes the known identities to the specified Writer
func (c *Config) EmitIdentities(out io.Writer) {
	EmitIdentityHeader(out)
	for _, identity := range c.Identities {
		EmitIdentity(out, identity)
	}
}
