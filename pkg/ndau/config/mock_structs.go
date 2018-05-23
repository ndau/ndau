package config

import "github.com/oneiro-ndev/msgp-well-known-types/wkt"

//go:generate msgp

//msgp:tuple MockKeyValue

type MockKeyValue struct {
	Key   wkt.Bytes
	Value wkt.Bytes
}

type MockNamespace struct {
	Namespace wkt.Bytes
	Data      []MockKeyValue
}

type MockChaosChain struct {
	Namespaces []MockNamespace
}
