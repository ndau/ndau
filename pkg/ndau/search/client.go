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

	"github.com/jackc/pgx/v4"
	metaapp "github.com/oneiro-ndev/metanode/pkg/meta/app"
	"github.com/pkg/errors"
)

// Client is a search Client that implements metaapp.Indexer.
type Client struct {
	// Postgres is a database connection
	Postgres *pgx.Conn

	// Used for getting account data to index.
	app AppIndexable

	// these values capture the position of a tx within a block
	height   uint64
	sequence uint64
}

// NewClient is a factory method for Client.
func NewClient(config *pgx.ConnConfig, app AppIndexable) (client *Client, err error) {
	client = &Client{
		app: app,
	}

	client.Postgres, err = pgx.ConnectConfig(context.Background(), config)
	err = errors.Wrap(err, "connecting to postgres")
	return
}

var _ metaapp.Indexer = (*Client)(nil)
