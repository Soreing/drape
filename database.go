package ssql

import (
	"context"
	"database/sql"
	"time"

	"github.com/Soreing/easyscan"
)

type DB struct {
	db         *sql.DB
	afterQuery QueryHook
}

// Connects to the database with no hook.
func Connect(
	ctx context.Context,
	driver string,
	dsn string,
) (*DB, error) {
	return ConnectWithHook(ctx, driver, dsn, nil)
}

// Opens a connection to the database and ensures that it is alive.
// Attaches a hook that is called after queries are executed.
func ConnectWithHook(
	ctx context.Context,
	driver string,
	dsn string,
	afterQuery QueryHook,
) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{
		db:         db,
		afterQuery: afterQuery,
	}, nil
}

// Gets a single record from the database with query string and params.
// AfterQuery hook is called after the query is returned. The row is scanned
// into an object that implements the ScanRow function.
func (db *DB) Get(
	ctx context.Context,
	dest easyscan.ScanOne,
	query string,
	params ...interface{},
) (err error) {
	qd := QueryDetails{
		StartTime: time.Now(),
		Query:     query,
		Params:    params,
	}
	defer func() {
		if db.afterQuery != nil {
			db.afterQuery(ctx, qd, err)
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
// AfterQuery hook is called after the query is returned. The rows are scanned
// into an object that implements the ScanAppendRow function.
func (db *DB) Select(
	ctx context.Context,
	dest easyscan.ScanMany,
	query string,
	params ...interface{},
) (err error) {
	qd := QueryDetails{
		StartTime: time.Now(),
		Query:     query,
		Params:    params,
	}
	defer func() {
		if db.afterQuery != nil {
			db.afterQuery(ctx, qd, err)
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

// Executes a query in the database with parameters. AfterQuery hook is called
// after the query is returned.
func (db *DB) Exec(
	ctx context.Context,
	query string,
	params ...interface{},
) (res sql.Result, err error) {
	qd := QueryDetails{
		StartTime: time.Now(),
		Query:     query,
		Params:    params,
	}
	defer func() {
		if db.afterQuery != nil {
			db.afterQuery(ctx, qd, err)
		}
	}()

	res, err = db.db.ExecContext(ctx, query, params...)
	return
}

// Begins a transaction with default options
func (db *DB) Begin(
	ctx context.Context,
) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx:         tx,
		afterQuery: db.afterQuery,
	}, nil
}

// Begins a transaction with custom options
func (db *DB) Beginx(
	ctx context.Context,
	opt *sql.TxOptions,
) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opt)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx:         tx,
		afterQuery: db.afterQuery,
	}, nil
}

