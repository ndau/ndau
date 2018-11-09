// Package cache provides a thread-safe cache of system variables
package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/oneiro-ndev/chaos/pkg/genesisfile"
	generator "github.com/oneiro-ndev/chaos_genesis/pkg/genesis.generator"
	"github.com/oneiro-ndev/system_vars/pkg/svi"
	"github.com/oneiro-ndev/writers/pkg/testwriter"
	"github.com/sirupsen/logrus"
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

func makecache(t *testing.T, latency, timeout time.Duration) SystemCache {
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

// implied in this test: timeouts are always errors, never panics
func TestSystemCacheTimeout(t *testing.T) {
	logger := logrus.New()
	logger.Out = testwriter.New(t)

	tcases := []struct {
		latency time.Duration
		timeout time.Duration
		wanterr bool
	}{
		{0 * time.Millisecond, 50 * time.Millisecond, false},
		{100 * time.Millisecond, 50 * time.Millisecond, true},
	}
	for idx, tcase := range tcases {
		t.Run(
			fmt.Sprintf("latency: %s; timeout: %s", tcase.latency, tcase.timeout),
			func(t *testing.T) {
				sc := makecache(t, tcase.latency, tcase.timeout)
				err := sc.Update(uint64(idx+1), logger)

				if tcase.wanterr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			},
		)
	}
}
