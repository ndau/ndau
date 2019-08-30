package sdk

import (
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/pkg/errors"
)

// EAIRate returns EAI rates given certain account states
//
// The address field is just to correlate request fields with response fields;
// account data is not checked.
func (c *Client) EAIRate(query ...routes.EAIRateRequest) (response []routes.EAIRateResponse, err error) {
	response = make([]routes.EAIRateResponse, 0)
	err = c.post(query, &response, c.URL("system/eai/rate"))
	err = errors.Wrap(err, "getting EAI rates from API")
	return
}
