package ssql

import (
	"context"
	"database/sql"
	"time"

	"github.com/Soreing/easyscan"
)

type Db struct {
	db    *sql.DB
	hooks []QueryHook
}

// Opens a connection to the database and ensures that it is alive.
func Connect(
	ctx context.Context,
	driver string,
	dsn string,
) (DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &Db{
		db:    db,
		hooks: []QueryHook{},
	}, nil
}

// Gets a single record from the database with query string and params.
// Hooks are called after the query is returned. The row is scanned
// into an object that implements the ScanRow function.
func (db *Db) Get(
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
		for _, hk := range db.hooks {
			hk(ctx, qd, err)
		}
	}()

	rows, err := db.db.QueryContext(ctx, query, params...)
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
func (db *Db) Select(
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
		for _, hk := range db.hooks {
			hk(ctx, qd, err)
		}
	}()

	rows, err := db.db.QueryContext(ctx, query, params...)
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
func (db *Db) Exec(
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
		for _, hk := range db.hooks {
			hk(ctx, qd, err)
		}
	}()

	res, err = db.db.ExecContext(ctx, query, params...)
	return
}

// Begins a transaction with default options.
func (db *Db) Begin(
	ctx context.Context,
) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Tx{
		db: db,
		tx: tx,
	}, nil
}

// Begins a transaction with custom options.
func (db *Db) Beginx(
	ctx context.Context,
	opt *sql.TxOptions,
) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opt)
	if err != nil {
		return nil, err
	}

	return &Tx{
		db: db,
		tx: tx,
	}, nil
}

// Adds a hook that is eecuted after the query is returned.
// Hooks are propagated to transactions started by the db context.
// The hook can not be removed.
func (db *Db) UseHook(
	fn func(context.Context, QueryDetails, error),
) {
	db.hooks = append(db.hooks, fn)
}
