// Package cache provides a thread-safe cache of system variables
package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/oneiro-ndev/chaos/pkg/genesisfile"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	"github.com/oneiro-ndev/system_vars/pkg/svi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tinylib/msgp/msgp"
)

// SystemCache is a thread-safe cache of system variables
type SystemCache struct {
	inner map[string][]byte
	lock  sync.RWMutex

	store   svi.SystemStore
	svi     svi.Location
	timeout time.Duration
}

// NewSystemCache constructs a SystemCache
func NewSystemCache(conf config.Config) (*SystemCache, error) {
	var ss svi.SystemStore
	var err error
	if conf.UseMock != nil && len(*conf.UseMock) > 0 {
		ss, err = genesisfile.Load(*conf.UseMock)
		if err != nil {
			return nil, errors.Wrap(err, "System() failed to load mock")
		}
	} else {
		ss = newChaosClient(conf.ChaosAddress)
	}

	return &SystemCache{
		inner:   make(map[string][]byte),
		lock:    sync.RWMutex{},
		store:   ss,
		svi:     conf.SystemVariableIndirect,
		timeout: time.Duration(conf.ChaosTimeout) * time.Millisecond,
	}, nil
}

type kv struct {
	k string
	v []byte
}

func (c *SystemCache) getKV(
	done <-chan struct{},
	name string, nsk svi.Location,
) (<-chan kv, <-chan error) {
	results := make(chan kv)
	errors := make(chan error)

	go func() {
		defer close(results)
		defer close(errors)

		value, err := c.store.GetRaw(nsk)
		select {
		case <-done:
			// return without sending anything; since we never write to
			// done, a successful read means that the channel is closed, which
			// means that the output channels must also be closed
		default:
			if err != nil {
				errors <- err
			} else {
				results <- kv{
					k: name,
					v: value,
				}
			}
		}
	}()

	return results, errors
}

// Update the cache
//
// height must be the current application height
// logger should be the system logger or nil
func (c *SystemCache) Update(height uint64, logger log.FieldLogger) error {
	// get a write lock and replace the inner map
	c.lock.Lock()
	defer c.lock.Unlock()

	logger = logger.WithField("height", height)
	logger.Info("SystemCache.Update started")

	// get the map of system variables
	// we get this each time instead of caching because
	// we want to stay updated in case the map is updated
	sviMap, err := svi.GetSVI(c.store, c.svi)
	if err != nil {
		return errors.Wrap(err, "could not get SVI map")
	}

	// actually getting the keys and values is super IO-heavy,
	// so we do it asynchronously
	//
	// here, we dispatch a bunch of goroutines, each of which is
	// responsible for fetching a single value for the requested key
	//
	// unfortunately, golang's channels _really_ don't like to play nicely
	// in a MPSC configuration, and are super easy to panic if you try, so
	// we have to faff around with some boilerplate to make this work right.
	//
	// done isn't actually a stream: we use it for simultaneous multi-channel
	// communication: if it's still open, we're still reading from the streams.
	done := make(chan struct{})
	defer close(done)

	// construct some lists of channels which we can populate, one per gorountine
	resultschannels := make([]<-chan kv, 0, len(sviMap))
	errorschannels := make([]<-chan error, 0, len(sviMap))

	// dispatch goroutines, collecting results channels in the process
	for name, dc := range sviMap {
		var nsk svi.Location
		if height >= dc.ChangeOn {
			nsk = dc.Future
		} else {
			nsk = dc.Current
		}
		r, e := c.getKV(done, name, nsk)
		resultschannels = append(resultschannels, r)
		errorschannels = append(errorschannels, e)
	}

	resultsStream := mergeKV(done, resultschannels...)
	errorsStream := mergeErr(done, errorschannels...)

	// create a new cache map.
	//
	// This ensures that even if this update fails,
	// we keep the results of the previous cache
	newCache := make(map[string][]byte)

	// create the timeout channel outside the loop so it's not reset on each
	// iteration
	timeout := time.After(c.timeout)

	// now collect the results: we know each goroutine will send once on either
	// the resultsStream or the errorsStream, so we can just use a static for
	// loop to collect the right number of results
	for i := 0; i < len(sviMap); i++ {
		var kv kv
		select {
		case kv = <-resultsStream:
			newCache[kv.k] = kv.v
		case err = <-errorsStream:
			return errors.Wrap(err, "could not get system variable "+kv.k)
		case <-timeout:
			return fmt.Errorf(
				"attempt to get system variables timed out: collected %d of %d values in %s",
				len(newCache),
				len(sviMap),
				c.timeout,
			)
		}
	}

	// log the sys variable keys available here
	// for debugging only
	// keys := make([]string, len(newCache))
	// i := 0
	// for k := range newCache {
	// 	keys[i] = k
	// 	i++
	// }
	// logger.WithField("system variable keys", keys).Info("SystemCache.Update completed")

	// everything's fine; just replace the inner cache with the new one now
	c.inner = newCache
	return nil
}

// GetRaw returns the raw bytes of the specified system variable
func (c *SystemCache) GetRaw(name string) []byte {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.inner[name]
}

// Get unmarshals the specified system variable into the provided value
func (c *SystemCache) Get(name string, value msgp.Unmarshaler) error {
	valBytes := c.GetRaw(name)
	if valBytes == nil {
		return fmt.Errorf("Requested system variable '%s' does not exist", name)
	}
	leftover, err := value.UnmarshalMsg(valBytes)
	if len(leftover) > 0 {
		return errors.New("Provided value type did not completely unmarshal system variable")
	}
	return err
}

// Set sets the specified system variable
//
// This is useful for overriding system variables for testing, but it's almost
// certainly not what you want in production. Think VERY HARD before using this
// method outside of a testing context.
func (c *SystemCache) Set(name string, value msgp.Marshaler) error {
	bytes, err := value.MarshalMsg(nil)
	if err != nil {
		return errors.Wrap(err, "marshalling value in set")
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	c.inner[name] = bytes
	return nil
}
