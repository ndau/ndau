package config

import "github.com/oneiro-ndev/msgp-well-known-types/wkt"

//go:generate msgp

//msgp:tuple MockKeyValue

// MockKeyValue mocks the chaos k-v pairs
type MockKeyValue struct {
	Key   wkt.Bytes
	Value wkt.Bytes
}

// MockNamespace mocks a chaos namespace
type MockNamespace struct {
	Namespace wkt.Bytes
	Data      []MockKeyValue
}

// MockChaosChain mocks the chaos chain, without history
type MockChaosChain struct {
	Namespaces []MockNamespace
}
