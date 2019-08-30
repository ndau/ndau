package tool

import (
	"fmt"
	"sync"

	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/rpc/client"
)

// NodeResponse represents a response from the nodes call.
type NodeResponse struct {
	Err   error
	Nodes []p2p.DefaultNodeInfo
}

// NodesClient is the interface required by the Nodes function
type NodesClient interface {
	client.StatusClient
	client.NetworkClient
}

// Nodes returns NodeInfo asyncronously.
func Nodes(node NodesClient) chan NodeResponse {
	var wg sync.WaitGroup
	wg.Add(2) // going to make 2 requests

	respCh := make(chan NodeResponse)

	// get self
	go func() {
		defer wg.Done()
		// get the node info from the node in the config
		status, err := Info(node)
		if err != nil {
			respCh <- NodeResponse{Err: fmt.Errorf("could not fetch node info: %v", err)}
			return
		}
		respCh <- NodeResponse{Nodes: []p2p.DefaultNodeInfo{status.NodeInfo}}
	}()

	// get the peer nodes and add their NodeInfo
	go func() {
		defer wg.Done()
		netInfo, err := node.NetInfo()
		if err != nil {
			respCh <- NodeResponse{Err: fmt.Errorf("could not fetch net info: %v", err)}
			return
		}
		var res []p2p.DefaultNodeInfo
		for _, p := range netInfo.Peers {
			res = append(res, p.NodeInfo)
		}
		respCh <- NodeResponse{Nodes: res}
	}()

	// wait until both requests are done, then close the channel
	go func() {
		wg.Wait()
		close(respCh)
	}()

	return respCh
}
