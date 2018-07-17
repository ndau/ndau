package cache

import (
	"github.com/oneiro-ndev/chaostool/pkg/tool"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"
	"github.com/pkg/errors"
	trpc "github.com/tendermint/tendermint/rpc/client"
	"github.com/tinylib/msgp/msgp"
)

type chaosClient struct {
	inner trpc.ABCIClient
}

// Static type assertion that chaosClient implements the SystemStore interface
var _ config.SystemStore = (*chaosClient)(nil)

func newChaosClient(address string) chaosClient {
	return chaosClient{
		inner: trpc.NewHTTP(address, "/websocket"),
	}
}

// GetRaw implements the SystemStore interface
func (cc chaosClient) GetRaw(namespace []byte, key msgp.Marshaler) ([]byte, error) {
	keyBytes, err := key.MarshalMsg([]byte{})
	if err != nil {
		return nil, errors.Wrap(err, "chaosClient.GetRaw failed to marshal key")
	}
	response, _, err := tool.GetNamespacedAt(cc.inner, namespace, keyBytes, 0)
	if err != nil {
		return nil, errors.Wrap(err, "chaosClient.GetRaw failed to get value")
	}
	return response, nil
}

// Get implements the SystemStore interface
func (cc chaosClient) Get(
	namespace []byte,
	key msgp.Marshaler,
	value msgp.Unmarshaler,
) error {
	return tool.GetStructured(
		cc.inner, namespace, key, value, 0,
	)
}
