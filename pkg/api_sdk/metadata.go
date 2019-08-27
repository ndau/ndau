package sdk

import (
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/pkg/errors"
)

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
