package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/oneiro-ndev/chaos/pkg/genesisfile"
	generator "github.com/oneiro-ndev/chaos_genesis/pkg/genesis.generator"
	"github.com/oneiro-ndev/system_vars/pkg/svi"
	"github.com/stretchr/testify/require"
	"github.com/tinylib/msgp/msgp"
)

type mockSystemStore struct {
	gfile   genesisfile.GFile
	latency time.Duration
}

func makeMockSystemStore(t *testing.T, latency time.Duration) mockSystemStore {
	_, gfilepath, _, err := generator.GenerateIn("")
	require.NoError(t, err)

	gfile, err := genesisfile.Load(gfilepath)
	require.NoError(t, err)

	return mockSystemStore{
		gfile:   gfile,
		latency: latency,
	}
}

var _ svi.SystemStore = (*mockSystemStore)(nil)

// Get implements svi.SystemStore
func (m mockSystemStore) Get(loc svi.Location, value msgp.Unmarshaler) error {
	time.Sleep(m.latency)
	return m.gfile.Get(loc, value)
}

// GetRaw implements svi.SystemStore
func (m mockSystemStore) GetRaw(loc svi.Location) ([]byte, error) {
	time.Sleep(m.latency)
	return m.gfile.GetRaw(loc)
}

// MakeMockCache makes a mock cache with specified latency and timeout for testing
func MakeMockCache(t *testing.T, latency, timeout time.Duration) SystemCache {
	ss := makeMockSystemStore(t, latency)
	sviLoc, err := ss.gfile.FindSVIStub()
	require.NoError(t, err)

	return SystemCache{
		inner:   make(map[string][]byte),
		lock:    sync.RWMutex{},
		store:   ss,
		svi:     *sviLoc,
		timeout: timeout,
	}
}
