package tool

import (
	"encoding/json"
	"io"
	"os"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/pkg/errors"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	ttypes "github.com/tendermint/tendermint/types"
)

// JSONClient conforms to the client.ABCIClient interface, but is not
// in fact an ABCI client. Instead, it marshals incoming transactions into
// their canonical JSON forms and emits them on the internal writer,
// defaulting to stdout
type JSONClient struct {
	Pretty bool
	Writer io.Writer
}

// NewJSONClient creates a new JSONClient
func NewJSONClient(pretty bool) JSONClient {
	return JSONClient{
		Pretty: pretty,
		Writer: os.Stdout,
	}
}

var _ client.ABCIClient = (*JSONClient)(nil)

// ABCIInfo implements ABCIClient
func (JSONClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	return nil, errors.New("ABCIInfo not implemented for JSONClient")
}

// ABCIQuery implements ABCIClient
func (JSONClient) ABCIQuery(path string, data cmn.HexBytes) (*ctypes.ResultABCIQuery, error) {
	return nil, errors.New("ABCIQuery not implemented for JSONClient")
}

// ABCIQueryWithOptions implements ABCIClient
func (JSONClient) ABCIQueryWithOptions(path string, data cmn.HexBytes, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	return nil, errors.New("ABCIQueryWithOptions not implemented for JSONClient")
}

// BroadcastTxCommit implements ABCIClient
func (j JSONClient) BroadcastTxCommit(tx ttypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	return nil, j.broadcast(tx)
}

// BroadcastTxSync implements ABCIClient
func (j JSONClient) BroadcastTxSync(tx ttypes.Tx) (*ctypes.ResultBroadcastTx, error) {
	return nil, j.broadcast(tx)
}

// BroadcastTxAsync implements ABCIClient
func (j JSONClient) BroadcastTxAsync(tx ttypes.Tx) (*ctypes.ResultBroadcastTx, error) {
	return nil, j.broadcast(tx)
}

// don't actually broadcast this tx
// instead, transform it into its canonical json representation and emit
func (j JSONClient) broadcast(txb ttypes.Tx) error {
	tx, err := metatx.Unmarshal(txb, ndau.TxIDs)
	if err != nil {
		return err
	}
	var jsonb []byte
	if j.Pretty {
		jsonb, err = json.MarshalIndent(tx, "", "  ")
	} else {
		jsonb, err = json.Marshal(tx)
	}
	if err != nil {
		return err
	}
	_, err = j.Writer.Write(jsonb)
	return err
}
