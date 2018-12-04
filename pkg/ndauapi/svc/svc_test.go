package svc

import (
	"net/http"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/stretchr/testify/assert"
)

func TestRouting(t *testing.T) {
	type rt struct {
		verb string
		in   string
		out  string
	}

	routes := []rt{
		rt{"GET", "/account/account/123456", "/account/account/:address"},
		rt{"POST", "/account/accounts", "/account/accounts"},
		rt{"POST", "/account/eai/rate", "/account/eai/rate"},
		rt{"GET", "/account/history/123456", "/account/history/:address"},
		rt{"GET", "/block/hash/abc123", "/block/hash/:blockhash"},
		rt{"GET", "/block/height/10234", "/block/height/:height"},
		rt{"GET", "/block/range/123/143", "/block/range/:first/:last"},
		rt{"GET", "/chaos/value/mynamespace/all", "/chaos/value/:namespace/all"},
		rt{"GET", "/chaos/value/mynamespace/TargetPrice", "/chaos/value/:namespace/:key"},
		rt{"GET", "/chaos/history/mynamespace/FloorPrice", "/chaos/history/:namespace/:key"},
		rt{"GET", "/node/status", "/node/status"},
		rt{"GET", "/node/health", "/node/health"},
		rt{"GET", "/node/net", "/node/net"},
		rt{"GET", "/node/genesis", "/node/genesis"},
		rt{"GET", "/node/abci", "/node/abci"},
		rt{"GET", "/node/consensus", "/node/consensus"},
		rt{"GET", "/node/nodes", "/node/nodes"},
		rt{"GET", "/node/ad349f", "/node/:id"},
		rt{"GET", "/order/hash/54589a34", "/order/hash/:ndauhash"},
		rt{"GET", "/order/height/10888", "/order/height/:ndauheight"},
		rt{"GET", "/order/history", "/order/history"},
		rt{"GET", "/order/current", "/order/current"},
		rt{"GET", "/transaction/5469abfed", "/transaction/:txhash"},
		rt{"POST", "/tx/prevalidate", "/tx/prevalidate"},
		rt{"POST", "/tx/submit", "/tx/submit"},
		rt{"GET", "/version", "/version"},
	}

	cf := cfg.Cfg{}
	mux := New(cf).Mux()

	for _, r := range routes {
		t.Run(r.out, func(t *testing.T) {
			req, _ := http.NewRequest(r.verb, r.in, nil)
			route := mux.GetRequestRoute(req)
			assert.Equal(t, r.out, route)
		})
	}
}
