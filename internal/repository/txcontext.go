package repository

import (
	"context"
)

// txKey in the internal key type for extracting a transaction
type txKey struct{}

// ctxWithTx returns an enriched context with a transaction object
// stored in
func ctxWithTx(ctx context.Context, tx Executor) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// executorFrom will verify if the context carries a transaction object
// and, if so, return it. Otherwise, will
func executorFrom(ctx context.Context, db Executor) Executor {
	// verify if the context carries a transaction object (which is an Executor type)
	tx, ok := ctx.Value(txKey{}).(Executor)
	// if not, return the regular non-transaction db connection
	if !ok {
		return db
	}
	return tx
}
