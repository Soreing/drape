package ssql

import (
	"context"
	"database/sql"
	"time"

	"github.com/Soreing/easyscan"
)

type Tx struct {
	db *Db
	tx *sql.Tx
}

// Gets a single record from the database with query string and params.
// Hooks are called after the query is returned. The row is scanned
// into an object that implements the ScanRow function.
func (tx *Tx) Get(
	ctx context.Context,
	dest easyscan.ScanOne,
	query string,
	params ...interface{},
) (err error) {
	qd := QueryDetails{
		StartTime: time.Now(),
		Function:  "Get",
		Query:     query,
		Params:    params,
	}
	defer func() {
		for _, hk := range tx.db.hooks {
			hk(ctx, qd, err)
		}
	}()

	rows, err := tx.tx.QueryContext(ctx, query, params...)
	if err != nil {
		return
	}

	defer rows.Close()
	if !rows.Next() {
		err = sql.ErrNoRows
	} else {
		err = dest.ScanRow(rows)
	}
	return
}

// Selects multiple records from the database with query string and params.
// Hooks are called after the query is returned. The rows are scanned
// into an object that implements the ScanAppendRow function.
func (tx *Tx) Select(
	ctx context.Context,
	dest easyscan.ScanMany,
	query string,
	params ...interface{},
) (err error) {
	qd := QueryDetails{
		StartTime: time.Now(),
		Function:  "Select",
		Query:     query,
		Params:    params,
	}
	defer func() {
		for _, hk := range tx.db.hooks {
			hk(ctx, qd, err)
		}
	}()

	rows, err := tx.tx.QueryContext(ctx, query, params...)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		err = dest.ScanAppendRow(rows)
		if err != nil {
			return
		}
	}
	return
}

// Executes a query in the database with parameters.
// Hooks are called after the query is returned.
func (tx *Tx) Exec(
	ctx context.Context,
	query string,
	params ...interface{},
) (res sql.Result, err error) {
	qd := QueryDetails{
		StartTime: time.Now(),
		Function:  "Exec",
		Query:     query,
		Params:    params,
	}
	defer func() {
		for _, hk := range tx.db.hooks {
			hk(ctx, qd, err)
		}
	}()

	res, err = tx.tx.ExecContext(ctx, query, params...)
	return
}

// Commits the transaction.
func (tx *Tx) Commit(
	ctx context.Context,
) error {
	err := tx.tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Rolls back the transaction.
func (tx *Tx) Rollback(
	ctx context.Context,
) error {
	err := tx.tx.Rollback()
	if err != nil {
		return err
	}
	return nil
}
