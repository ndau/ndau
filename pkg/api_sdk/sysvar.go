package sdk

import (
	"strings"

	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// Sysvars gets the list of requested system variables, marshaled
func (c *Client) Sysvars(vars ...string) (svs map[string][]byte, err error) {
	var url string
	if len(vars) == 0 {
		url = c.URL("system/all")
	} else {
		url = c.URL("system/get/%s", strings.Join(vars, ","))
	}
	err = c.get(svs, url)
	err = errors.Wrap(err, "getting sysvars")
	return
}

// Sysvars gets the list of requested system variables, marshaled
func Sysvars(node *Client, vars ...string) (map[string][]byte, error) {
	return node.Sysvars(vars...)
}

// Sysvar gets a single system variable given its name and an example of its type
//
// The example is populated with the appropriate data
func (c *Client) Sysvar(name string, example msgp.Unmarshaler) error {
	svs, err := c.Sysvars(name)
	if err != nil {
		return errors.Wrap(err, "getting system variable")
	}
	sv, ok := svs[name]
	if !ok {
		return errors.New("API did not return requested system variable")
	}
	_, err = example.UnmarshalMsg(sv)
	return errors.Wrap(err, "unmarshaling")
}

// Sysvar gets a single system variable given its name and an example of its type
//
// The example is populated with the appropriate data
func Sysvar(node *Client, name string, example msgp.Unmarshaler) error {
	return node.Sysvar(name, example)
}

// SysvarHistory gets the value history of the given sysvar.
//
// Pass in 0,0 for the paging params to get the entire history.
func (c *Client) SysvarHistory(name string, after uint64, limit int) (resp *query.SysvarHistoryResponse, err error) {
	err = c.get(resp, c.URLP(
		params{"after": after, "limit": limit},
		"system/history/%s", name,
	))
	err = errors.Wrap(err, "getting sysvar history")
	return
}

// SysvarHistory gets the value history of the given sysvar.
//
// Pass in 0,0 for the paging params to get the entire history.
func SysvarHistory(node *Client, name string, after uint64, limit int) (*query.SysvarHistoryResponse, error) {
	return node.SysvarHistory(name, after, limit)
}
