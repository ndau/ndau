package tool

import (
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tinylib/msgp/msgp"
)

// Sysvars gets the version the connected node is running
func Sysvars(node client.ABCIClient, vars ...string) (
	map[string][]byte, *rpctypes.ResultABCIQuery, error,
) {
	var rqb []byte
	var err error
	if len(vars) > 0 {
		rqb, err = query.SysvarsRequest(vars).MarshalMsg(nil)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to marshal sysvar request")
		}
	}
	// perform the query
	res, err := node.ABCIQuery(query.SysvarsEndpoint, rqb)
	if err != nil {
		return nil, res, err
	}
	resp := make(query.SysvarsResponse)
	_, err = resp.UnmarshalMsg(res.Response.Value)
	return resp, res, err
}

// Sysvar gets a single system variable given its name and an example of its type
//
// The example is populated with the appropriate data
func Sysvar(node client.ABCIClient, name string, example msgp.Unmarshaler) error {
	svs, _, err := Sysvars(node, name)
	if err != nil {
		return err
	}
	svb, ok := svs[name]
	if !ok {
		return errors.New("node did not return sysvar " + name)
	}
	_, err = example.UnmarshalMsg(svb)
	return err
}
