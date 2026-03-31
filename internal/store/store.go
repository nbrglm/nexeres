// Package store  provides a PostgreSQL database connection and utility functions.
// It includes functions for initializing the database connection pool, closing it, etc.
//
// It is used to keep a global connection to the PostgreSQL database,
// which can be used by other packages in the application to interact with the database.
package store

import (
	"context"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/opts"
)

// A DB Connection Pool
var PgPool *pgxpool.Pool

// A Querier for executing queries
var Querier *db.Queries

// InitDB initializes the database connection pool.
//
// It should be called at the start of the application to set up the database connection pool.
func InitDB(ctx context.Context) (err error) {
	pgConfig, err := pgxpool.ParseConfig(config.C.Stores.PostgresDSN)
	if err != nil {
		return err
	}

	options := []otelpgx.Option{
		otelpgx.WithTrimSQLInSpanName(),
	}

	if !opts.Debug {
		options = append(options, otelpgx.WithDisableSQLStatementInAttributes())
	}

	// Add Tracer
	pgConfig.ConnConfig.Tracer = otelpgx.NewTracer(options...)

	PgPool, err = pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		return err
	}

	if err := otelpgx.RecordStats(PgPool); err != nil {
		return fmt.Errorf("unable to record database stats: %w", err)
	}

	Querier = db.New(PgPool)
	return nil
}

// Ping pings the database connections in the pool and returns any failures
//
// Can be used to health check
func Ping(ctx context.Context) error {
	return PgPool.Ping(ctx)
}

// CloseDB closes the database connection pool.
//
// It should be called at the end of the application to clean up the database connection pool.
// It closes the global `PgPool`. It is a blocking-call.
func CloseDB() error {
	if PgPool != nil {
		PgPool.Close()
	}
	return nil
}
