package sdk_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"testing"
	"time"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	sdk "github.com/oneiro-ndev/ndau/pkg/api_sdk"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/mock"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

const (
	apiport = 4096
	sysvar  = "ReleaseFromEndowmentAddress"
)

func makeAddress(t *testing.T) address.Address {
	seed, err := key.GenerateSeed(key.RecommendedSeedLen)
	require.NoError(t, err)
	ekey, err := key.NewMaster(seed)
	require.NoError(t, err)
	ekeypub, err := ekey.Public()
	require.NoError(t, err)
	pubkeyi, err := ekeypub.AsSignatureKey()
	require.NoError(t, err)
	pubkey := pubkeyi.(*signature.PublicKey)
	addr, err := address.Generate(address.KindUser, pubkey.KeyBytes())
	require.NoError(t, err)
	return addr
}

func setup(t *testing.T, test func(*sdk.Client), initAddrs ...address.Address) {
	cf := mock.Cfg(t, func(abapp abcitypes.Application) {
		app := abapp.(*ndau.App)
		app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
			state := stI.(*backing.State)
			var err error
			state.Sysvars[sysvar], err = makeAddress(t).MarshalMsg(nil)
			require.NoError(t, err)
			for n, addr := range initAddrs {
				state.Accounts[addr.String()] = backing.AccountData{
					Balance:  math.Ndau(100*n + 1),
					Sequence: uint64(n),
				}
			}
			return state, nil
		})
	})

	port := apiport + rand.Intn(1024)

	cf.Port = port

	ports := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:    ports,
		Handler: svc.NewLogMux(cf),
	}
	// the listener may take a moment to spin up, but this call blocks until
	// it's ready. From there, the server should be ready the moment it calls Serve.
	listener, err := net.Listen("tcp", ports)
	require.NoError(t, err)
	go server.Serve(listener)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	defer server.Shutdown(ctx)

	test(sdk.TestClient(t, uint(port)))
}

func TestTestSetupWorks(t *testing.T) {
	setup(t, func(*sdk.Client) {})
}

func TestABCIInfo(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.ABCIInfo()
		require.NoError(t, err)
	})
}

func TestConsensus(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.Consensus()
		require.NoError(t, err)
	})
}

func TestEAIRate(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.EAIRate(routes.EAIRateRequest{
			Address: "foo",
			WAA:     5 * math.Month,
		})
		require.NoError(t, err)
	})
}

func TestGenesis(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.Genesis()
		require.NoError(t, err)
	})
}

func TestGetAccount(t *testing.T) {
	addr := makeAddress(t)
	setup(t, func(client *sdk.Client) {
		_, err := client.GetAccount(addr)
		require.NoError(t, err)
	}, addr)
}

func TestGetAccountHistory(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.GetAccountHistory(search.AccountHistoryParams{
			Address: makeAddress(t).String(),
		})
		require.NoError(t, err)
	})
}

func TestGetAccountList(t *testing.T) {
	const count = 10
	accts := make([]address.Address, 0, count)
	for i := 0; i < count; i++ {
		accts = append(accts, makeAddress(t))
	}
	setup(t, func(client *sdk.Client) {
		_, err := client.GetAccountList("", 0)
		require.NoError(t, err)
	}, accts...)
}

func TestGetAccountListBatch(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.GetAccountListBatch()
		require.NoError(t, err)
	})
}

func TestGetBlock(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.GetBlock("")
		require.NoError(t, err)
	})
}

func TestGetBlockAt(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.GetBlockAt(0)
		require.NoError(t, err)
	})
}

func TestGetBlocksByDaterange(t *testing.T) {
	last := time.Now()
	first := last.Add(-30 * 24 * time.Hour)
	setup(t, func(client *sdk.Client) {
		_, err := client.GetBlocksByDaterange(first, last, false, first, 0)
		require.NoError(t, err)
	})
}

func TestGetBlocksByHeight(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.GetBlocksByHeight(100, 200, false)
		require.NoError(t, err)
	})
}

func TestGetBlocksByRange(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.GetBlocksByRange(100, 200, false)
		require.NoError(t, err)
	})
}

func TestGetCurrencySeats(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.GetCurrencySeats()
		require.NoError(t, err)
	})
}

func TestGetCurrentBlock(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.GetCurrentBlock()
		require.NoError(t, err)
	})
}

func TestGetDelegates(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.GetDelegates()
		require.NoError(t, err)
	})
}

func TestGetSequence(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.GetSequence(makeAddress(t))
		require.NoError(t, err)
	})
}

func TestHealth(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.Health()
		require.NoError(t, err)
	})
}

func TestInfo(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.Info()
		require.NoError(t, err)
	})
}

func TestNetInfo(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.NetInfo()
		require.NoError(t, err)
	})
}

func TestPrevalidate(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, _, err := client.Prevalidate(ndau.NewIssue(1, 1))
		require.NoError(t, err)
	})
}

func TestPriceAt(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.PriceAt(1)
		require.NoError(t, err)
	})
}

func TestPriceInfo(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.PriceInfo()
		require.NoError(t, err)
	})
}

func TestSend(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.Send(ndau.NewIssue(1, 1))
		require.NoError(t, err)
	})
}

func TestSysvar(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		addr := address.Address{}
		err := client.Sysvar(sysvar, &addr)
		require.NoError(t, err)
	})
}

func TestSysvarHistory(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.SysvarHistory(sysvar, 0, 0)
		require.NoError(t, err)
	})
}

func TestSysvars(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.Sysvars()
		require.NoError(t, err)
	})
}

func TestVersion(t *testing.T) {
	setup(t, func(client *sdk.Client) {
		_, err := client.Version()
		require.NoError(t, err)
	})
}
