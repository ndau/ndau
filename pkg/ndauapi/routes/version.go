package routes

import (
	"net/http"
	"regexp"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

// VersionResult is returned from the /version request; it retrieves
// information from both the ndau and chaos chains
type VersionResult struct {
	ChaosVersion string
	ChaosSha     string
	NdauVersion  string
	NdauSha      string
	Network      string
}

// HandleVersion is an http handler for version info
func HandleVersion(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// first find a node to talk to
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving ndau node", err, http.StatusInternalServerError))
			return
		}

		version, _, err := tool.Version(node)
		if version == "" {
			version = "unidentified-0-gunknown"
		}

		// restore this when we get versioning into chaos
		// chnode, err := ws.Node(cf.ChaosAddress)
		// if err != nil {
		// 	reqres.RespondJSON(w, reqres.NewFromErr("error retrieving chaos node", err, http.StatusInternalServerError))
		// 	return
		// }

		// systemkey, _ := base64.StdEncoding.DecodeString("zBQ176aLnfZLZVugxik0T4p+t3RLG6AeDXDWoHdJEVY=")
		// systemkey := cns.System
		// resp, _, err := chtool.DumpNamespacedAt(chnode, systemkey, 0)
		// fmt.Println(resp, err)

		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving version info", err, http.StatusInternalServerError))
			return
		}

		// The NetworkName system variable is intended to be set to something like "ndau MainNet"
		// once the mainnet is live. It's expected to be used to identify the network
		// that the API is talking to, as a way to differentiate the ndau mainnet from
		// test networks.
		//
		// However, if it's not specified, we assume it's devnet.
		networkName := "ndau devnet"
		sysvars, err := getSystemVars(cf.NodeAddress)
		if err == nil {
			if n, ok := sysvars["NetworkName"]; ok {
				networkName = string(n)
			}
		}

		result := VersionResult{
			NdauVersion: version,
			Network:     networkName,
		}

		// Our default version format is generated by "git describe --long --tags"
		// which looks like "v0.7.8-23-g7c8eac5", where the 3 parts separated by dashes
		// are the version tag, the number of commits since then, and "g" followed by the
		// current commit hash.
		// We don't care about the number of commits, and the g is not interesting, so
		// we pattern match the middle part.
		// If the format doesn't match that, we will just return the version string unmodified.
		p := regexp.MustCompile("-[0-9]+-g")
		spv := p.Split(version, -1)
		result.NdauVersion = spv[0]
		if len(spv) > 1 {
			result.NdauSha = spv[len(spv)-1]
		}

		reqres.RespondJSON(w, reqres.Response{Bd: result, Sts: http.StatusOK})
	}
}
