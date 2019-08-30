package sdk

import (
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/pkg/errors"
)

// PriceInfo returns current price data for key parameters
func (c *Client) PriceInfo() (info *routes.PriceInfo, err error) {
	info = new(routes.PriceInfo)
	err = c.get(info, c.URL("price/current"))
	err = errors.Wrap(err, "fetching price info from API")
	return
}

// PriceAt returns price data at the given height
func (c *Client) PriceAt(height uint64) (info *routes.PriceInfo, err error) {
	info = new(routes.PriceInfo)
	err = c.get(info, c.URL("price/height/%d", height))
	err = errors.Wrap(err, "fetching price data from API")
	return
}
