package sdk

import (
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/pkg/errors"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Info gets the node's current status
func (c *Client) Info() (status *rpctypes.ResultStatus, err error) {
	err = c.get(status, c.URL("node/status"))
	err = errors.Wrap(err, "fetching node status from API")
	return
}

// Info gets the node's current status
func Info(node *Client) (status *rpctypes.ResultStatus, err error) {
	return node.Info()
}

// PriceInfo returns current price data for key parameters
func (c *Client) PriceInfo() (info *routes.PriceInfo, err error) {
	err = c.get(info, c.URL("price/current"))
	err = errors.Wrap(err, "fetching price info from API")
	return
}

// GetSIB exists for compatibility; it delegates to node.PriceInfo()
func GetSIB(node *Client) (*routes.PriceInfo, error) {
	return node.PriceInfo()
}

// GetSummary exists for compatibility; it delegates to node.PriceInfo()
func GetSummary(node *Client) (*routes.PriceInfo, error) {
	return node.PriceInfo()
}

// Version delivers version information
func (c *Client) Version() (version *routes.VersionResult, err error) {
	err = c.get(version, c.URL("version"))
	err = errors.Wrap(err, "getting version from API")
	return
}

// Version delivers version information
func Version(node *Client) (*routes.VersionResult, error) {
	return node.Version()
}
