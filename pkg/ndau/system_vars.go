package ndau

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"encoding/base64"
	"fmt"

	"github.com/ndau/ndau/pkg/ndau/backing"
	log "github.com/sirupsen/logrus"
	"github.com/tinylib/msgp/msgp"
)

// System retrieves a named system variable.
func (app *App) System(name string, value msgp.Unmarshaler) (err error) {
	state := app.GetState().(*backing.State)
	bytes, exists := state.Sysvars[name]
	if !exists {
		return fmt.Errorf("Sysvar %s does not exist", name)
	}
	var leftovers []byte
	leftovers, err = value.UnmarshalMsg(bytes)
	if err == nil && len(leftovers) > 0 {
		err = fmt.Errorf("Sysvar %s has extra trailing bytes; this is suspicious", name)
	}
	if err != nil {
		app.DecoratedLogger().WithError(err).WithFields(log.Fields{
			"Sysvar.Name":  name,
			"Sysvar.Value": base64.StdEncoding.EncodeToString(bytes),
		}).Error("problem getting sysvar")
	}
	return
}
