// Package cache provides a thread-safe cache of system variables
package cache

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tinylib/msgp/msgp"

	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	"github.com/pkg/errors"
)

// SystemCache is a thread-safe cache of system variables
type SystemCache struct {
	inner map[string][]byte
	lock  sync.RWMutex

	store   config.SystemStore
	svi     config.NamespacedKey
	timeout time.Duration
}

// NewSystemCache constructs a SystemCache
func NewSystemCache(conf config.Config) (*SystemCache, error) {
	var ss config.SystemStore
	var err error
	if len(conf.UseMock) > 0 {
		ss, err = config.LoadMock(conf.UseMock)
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
	name string, nsk config.NamespacedKey,
	results chan<- kv, errors chan<- error,
) {
	value, err := c.store.GetRaw(nsk.Namespace.Bytes(), nsk.Key)
	if err != nil {
		errors <- err
	} else {
		results <- kv{
			k: name,
			v: value,
		}
	}
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
	sviMap, err := config.GetSVI(c.store, c.svi)
	if err != nil {
		return errors.Wrap(err, "could not get SVI map")
	}

	// set up some channels
	// we don't buffer them; the block time on reads should be trivial
	// compared to the network lag
	resultsStream := make(chan kv)
	defer close(resultsStream)
	errorsStream := make(chan error)
	defer close(errorsStream)

	// actually getting the keys and values is super IO-heavy,
	// so we do it asynchronously
	//
	// here, we dispatch a bunch of goroutines, each of which is
	// responsible for fetching a single value for the requested key
	//
	// the results and any errors are sent along the channels we set up earlier
	for name, dc := range sviMap {
		var nsk config.NamespacedKey
		if height >= dc.ChangeOn {
			nsk = dc.Future
		} else {
			nsk = dc.Current
		}
		go c.getKV(name, nsk, resultsStream, errorsStream)
	}

	// create a new cache map.
	//
	// This ensures that even if this update fails,
	// we keep the results of the previous cache
	newCache := make(map[string][]byte)

	// now collect the results: we know we should get one result from each
	// goroutine, and we had one of those per key in sviMap
	for i := 0; i < len(sviMap); i++ {
		var kv kv
		select {
		case kv = <-resultsStream:
			newCache[kv.k] = kv.v
		case err = <-errorsStream:
			return errors.Wrap(err, "could not get system variable from chaos chain")
		case <-time.After(c.timeout):
			return errors.New("Attempt to get system variables from chaos chain timed out")
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
