package cache

import (
	"github.com/oneiro-ndev/chaos/pkg/tool"
	"github.com/oneiro-ndev/system_vars/pkg/svi"
	"github.com/pkg/errors"
	trpc "github.com/tendermint/tendermint/rpc/client"
	"github.com/tinylib/msgp/msgp"
)

type chaosClient struct {
	inner trpc.ABCIClient
}

// Static type assertion that chaosClient implements the SystemStore interface
var _ svi.SystemStore = (*chaosClient)(nil)

func newChaosClient(address string) chaosClient {
	return chaosClient{
		inner: trpc.NewHTTP(address, "/websocket"),
	}
}

// GetRaw implements the SystemStore interface
func (cc chaosClient) GetRaw(loc svi.Location) ([]byte, error) {
	response, _, err := tool.GetNamespacedAt(cc.inner, loc.Namespace, loc.Key, 0)
	if err != nil {
		return nil, errors.Wrap(err, "chaosClient.GetRaw failed to get value")
	}
	return response, nil
}

// Get implements the SystemStore interface
func (cc chaosClient) Get(
	loc svi.Location,
	value msgp.Unmarshaler,
) error {
	bytes, err := cc.GetRaw(loc)
	if err != nil {
		return errors.Wrap(err, "chaosClient.Get")
	}
	_, err = value.UnmarshalMsg(bytes)
	return errors.Wrap(err, "unmarshalling chaos bytes")
}
