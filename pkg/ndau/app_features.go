package ndau

import (
	"os"
	"strconv"

	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
)

// Supported ndau features.
// Add more feature constants here as needed.
const (
	NoExchangeHoldsOnTransfer     meta.Feature = 0
	NoKeysOnSetValidation         meta.Feature = 1
	ResetUncreditedEAIOnCreditEAI meta.Feature = 2
)

// Helper function that sets up the ndau feature height gates.
// This should be the last thing in this file that needs to be modified when we add new features.
//
// All new features are turned on at the given future height, gotten from an environment var.
// We control it there instead of hard-coding a future height here, so that we can make a
// last-minute choice as to which height in the future we want to activate the new features
// that have been implemented since the last upgrade of mainnet.  This avoids having to make
// code changes and rebuild/stage/test/etc when choosing this height at mainnet upgrade time.
// It's hard to choose such a height at dev time, since it sometimes takes weeks to land, and in
// the mean time, mainnet keeps growing.
func assignFeatureHeights(futureHeight uint64) map[meta.Feature]uint64 {
	// How to modify the existing heights (and add new heights) below:
	//
	// When we add more features, we need to replace all features gated by futureHeight below
	// with the value we used in the environment variable the last time we upgraded mainnet.
	// Then the new features can be gated by the new futureHeight.  We'll then update the
	// environment vairable when we're ready to upgrade mainnet again.
	features := make(map[meta.Feature]uint64)
	features[NoExchangeHoldsOnTransfer]     = futureHeight
	features[NoKeysOnSetValidation]         = futureHeight
	features[ResetUncreditedEAIOnCreditEAI] = futureHeight
	return features
}

// Features allows us to gate ndau-specific feature logic based on app state.
type Features struct {
	meta.Features

	// The app that ndau features consult to know whether they are active at the current height.
	app *App

	// Map whose keys are features,
	// and values are the mainnet block height at which the feature becomes active.
	features map[meta.Feature]uint64
}

// InitFeatures sets up a new Features object with the given app.
func (app *App) InitFeatures() {
	// This environment variable
	newFeatureHeight := os.Getenv("NEW_FEATURE_HEIGHT")
	maxHeight, err := strconv.ParseUint(newFeatureHeight, 10, 64)
	if err != nil {
		// If the environment variable isn't set, or if it's not valid, use height 0
		// so that all features are active all the time.
		maxHeight = 0
	}

	features := assignFeatureHeights(maxHeight)

	// There should be no height larger than the max height.  This is useful for networks
	// that have reset their blockchains. They can use "0" for the environment variable.
	// This is a no-op under normal circumstances.
	for f, h := range features {
		if h > maxHeight {
			features[f] = maxHeight
		}
	}

	app.SetFeatures(
		&Features{
			app:      app,
			features: features,
		},
	)
}

// IsActive returns whether the given feature is currently active in the app.
func (f *Features) IsActive(feature meta.Feature) bool {
	// If features is nil, it means that all features are active all the time.
	if f.features == nil {
		return true
	}

	gateHeight, ok := f.features[feature]

	// Unknown features are always active by default.
	if !ok {
		return true
	}

	return f.app.Height() >= gateHeight
}
