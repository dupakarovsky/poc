package foundationpostgres

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool defines a new connection pool for serving connections
// to a postgres using the pgxpool driver
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {

	// build the database URI string for connecting to the database
	// eg: postgres://user:pass@localhost:5432/database?search_path=test&sslmode=require
	connString := buildConnectionString(cfg)

	// construct the connection pool
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parsing pool config: %w", err)
	}

	// Only override ParseConfig's own sane built-in defaults when cfg
	// actually specifies something — unconditionally assigning these from
	// a Config that leaves them at Go's zero value. pgxpool treated every
	// connection as instantly "too idle" and evicts it in a loop,
	// surfacing as a generic, unhelpful "too many failed attempts
	// acquiring connection" error with no obvious cause. Building it
	// correctly from the start this time, now that the failure mode is
	// known ahead of need.
	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}

	// create the connection pool form the config
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	// pool created. Define a timed context for pinging
	// the database
	poolCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// ping database
	err = pool.Ping(poolCtx)
	if err != nil {
		// close the connection pool before exiting
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	// ping successful. return the connectionp pool

	return pool, nil
}

//-----------------------------------------------------------------------------

func buildConnectionString(cfg Config) string {

	// map or key/value pairs for URL
	query := make(url.Values)

	// set the SLLMode in advance
	query.Set("sslmode", cfg.SSLMode.String())

	// set the URI search_path if a schema is provided
	if cfg.Schema != "" {
		query.Set("search_path", cfg.Schema)
	}

	// construct the URI for the connection command
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:     cfg.Database,
		RawQuery: query.Encode(), // encode the remaining params.
	}

	return u.String()
}
