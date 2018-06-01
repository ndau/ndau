# `node.address`: a unique public identifier for the ndau chain

Features from the [design whitepaper](https://github.com/oneiro-ndev/whitepapers/blob/master/ndau_design/addresses.md):

> ## Required Features
> - Related to a public key in a way that permits the holder of the corresponding private key to conduct transactions
> - Short enough to type if necessary
> - Has features (check digits, etc) to make it obvious when typographic errors have been made
> - Can be identified as an ndau address by inspection

An additional requirement in order to support HD wallets is that addresses contain an opaque binary blob: when generating a public key,

## Design

An `Address` is a serialized representation of an `address.Data` struct. Programs are encourad to use `address.Data` structs internally, and convert to and from `Address` objects at the user boundary.

An address has the following fields:

- `Key`: a public key
- `KeyAlgorithm`: the algorithm with which the key is generated and validated
- `Derivation`: a blob of data useful for HD wallets
- `CRC`: a CRC-32 hash of the rest of the fields
