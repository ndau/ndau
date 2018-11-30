// Package cache provides a thread-safe cache of system variables
package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/oneiro-ndev/writers/pkg/testwriter"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// implied in this test: timeouts are always errors, never panics
func TestSystemCacheTimeout(t *testing.T) {
	logger := logrus.New()
	logger.Out = testwriter.New(t)

	tcases := []struct {
		latency time.Duration
		timeout time.Duration
		wanterr bool
	}{
		{30 * time.Millisecond, 50 * time.Millisecond, false},
		{55 * time.Millisecond, 50 * time.Millisecond, true},
		{105 * time.Millisecond, 50 * time.Millisecond, true},
	}
	for idx, tcase := range tcases {
		t.Run(
			fmt.Sprintf("latency: %5s; timeout: %5s", tcase.latency, tcase.timeout),
			func(t *testing.T) {
				sc := MakeMockCache(t, tcase.latency, tcase.timeout)
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
