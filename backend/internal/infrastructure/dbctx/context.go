// SPDX-License-Identifier: AGPL-3.0-or-later

// Package dbctx provides context helpers for database transactions.
// It enables RLS (Row Level Security) by storing transactions in context.Context,
// allowing repositories to transparently use either a transaction or raw DB connection.
package dbctx

import (
	"context"
	"database/sql"
)

// Querier is a common interface for *sql.DB and *sql.Tx.
// It allows repositories to work transparently with either a raw DB connection
// or a transaction, enabling RLS isolation via transactional set_config.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Compile-time interface checks
var (
	_ Querier = (*sql.DB)(nil)
	_ Querier = (*sql.Tx)(nil)
)

// txKey is the context key for storing the current transaction.
type txKey struct{}

// WithTx returns a new context containing the given transaction.
// This is used by the RLS middleware to propagate the transaction
// through the request lifecycle.
func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext extracts the transaction from the context if present.
// Returns nil if no transaction is stored in the context.
func TxFromContext(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return nil
}

// GetQuerier returns the Querier to use for database operations.
// If a transaction is present in the context (set by RLS middleware),
// it returns the transaction. Otherwise, it returns the raw DB connection.
//
// This allows repositories to transparently benefit from RLS isolation
// when called within a transactional context, while still working
// correctly for operations that bypass RLS (e.g., migrations, admin tasks).
func GetQuerier(ctx context.Context, db *sql.DB) Querier {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return db
}
