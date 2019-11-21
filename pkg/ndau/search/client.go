package search

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"context"

	"github.com/jackc/pgx"
	metaapp "github.com/oneiro-ndev/metanode/pkg/meta/app"
	"github.com/pkg/errors"
)

// Client is a search Client that implements metaapp.Indexer.
type Client struct {
	// postgres is a database connection
	// If the appropriate fields are configured in the configuration, this will be
	// populated at initialization. Otherwise, it will be nil.
	// Therefore, all usages have to take the possibility of nility into consideration.
	postgres *pgx.Conn

	// Used for getting account data to index.
	app AppIndexable

	height uint64
}

// NewClient is a factory method for Client.
func NewClient(config *pgx.ConnConfig, app AppIndexable) (client *Client, err error) {
	client = &Client{
		app: app,
	}

	client.postgres, err = pgx.ConnectConfig(context.Background(), config)
	err = errors.Wrap(err, "connecting to postgres")
	return
}

var _ metaapp.Indexer = (*Client)(nil)
