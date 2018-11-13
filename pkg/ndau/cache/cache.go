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

// What we want here is to be able to spawn a bunch of query goroutines, each of which could
// possibly fail, and then return a collection of the results; if there are any errors,
// we will return one of them.

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
	k   string
	v   []byte
	err error
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
	// done isn't actually a stream: we use it for simultaneous multi-channel
	// communication: if it's still open, we're still reading from the streams.
	done := make(chan struct{})
	defer close(done)

	// construct a list of channels which we can populate, one per gorountine
	resultschannels := make([]<-chan kv, 0, len(sviMap))

	for name, dc := range sviMap {
		var nsk svi.Location
		if height >= dc.ChangeOn {
			nsk = dc.Future
		} else {
			nsk = dc.Current
		}

		// make a channel for this goroutine
		results := make(chan kv)
		// and record it in the result list
		resultschannels = append(resultschannels, results)
		// now spawn a goroutine to retrieve one result and stuff it into the results channel
		// it also monitors the done channel to know if it should bail out early
		// it closes the results channel in either case
		// we pass in name and nsk so that they're bound to this goroutine
		go func(results chan kv, name string, nsk svi.Location) {
			value, err := c.store.GetRaw(nsk)
			result := kv{k: name, v: value}
			if err != nil {
				result = kv{err: err}
			}
			select {
			case <-done:
				// return without sending anything; since we never write to
				// done, a successful read means that the channel is closed
			case results <- result:
				// feed back the result; if err is non-nil, then it should be logged and the value ignored
			}
			// no matter what, we want to close our channel
			close(results)
		}(results, name, nsk)
	}

	// it is probably overkill to do a general channel merge since each one of these
	// channels either errors or delivers a value, but it does defend against
	// slowing everything down because one query is slow.
	resultsStream := mergeKV(done, resultschannels...)

	// create a new cache map.
	//
	// This ensures that even if this update fails,
	// we keep the results of the previous cache
	newCache := make(map[string][]byte)

	// create the timeout channel outside the loop so it's not reset on each
	// iteration
	timeout := time.After(c.timeout)

	// now collect the results:
	// we can't know how many iterations we'll get and we want to have a timeout option,
	// so we have to break the loop manually
outer:
	for {
		select {
		case kv, real := <-resultsStream:
			if real {
				if kv.err != nil {
					return errors.Wrap(kv.err, "could not get system variable "+kv.k)
				}
				newCache[kv.k] = kv.v
			} else {
				// we're done, there are no more values to accumulate
				// did we get all of them?
				if len(newCache) != len(sviMap) {
					return fmt.Errorf("some system variables were not received: collected %d of %d values in %s",
						len(newCache),
						len(sviMap),
						c.timeout,
					)
				}
				// in either case, if real was false, we're done here.
				break outer
			}
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
	// fmt.Printf("sv keys:\n%#v\n", keys)

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
