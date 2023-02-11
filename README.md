# Simple SQL
Simple SQL is a light wrapper around the native sql library, using easyscan to scan rows into objects and implementing hooks for additional logic around queries.

## Usage
Get a Database object through connection. The connection includes a ping to ensure that the opened connection is active. The examples use a PostgreSQL database.
```golang
dsn := "host=127.0.0.1 port=5432 user=pg password=admin dbname=library"
db, err := ssql.Connect(context.TODO(), "postgres", dsn)
if err != nil {
	panic(err)
}
```

There are only a handful of functions for interacting with the database to keep it simple. All functions require a context, and queries which return rows require a destination object that implement `easyscan.ScanOne` or `easyscan.ScanMany`. Queries require a query string and a list of optional parameters.
```golang
query := `INSERT INTO books(id, title) VALUES($1, $2)`
res, err := db.Exec(context.TODO(), query, 1, "Beards and Beer")
```
```golang
bk := Book{}
err := db.Get(context.TODO(), &bk, "SELECT * FROM books WHERE id=$1", 1)
```
```golang
bkl := BookList{}
err := db.Select(context.TODO(), &bkl, "SELECT * FROM books")
```
For atomic operations with multiple queries, you can start a transaction. Transactions can be committed when finished or rolled back on error.
```golang
tx, err := db.Begin(context.TODO())
```
```golang
query := `UPDATE books SET title=$2 WHERE id=$1`
res, err := db.Exec(context.TODO(), query, 1, "The Ancient Earth")
if err != nil {
    tx.Rollback(context.TODO())
}
tx.Commit(context.TODO())
```

## Hooks
You can attach hooks to databases which will be executed after the queries have been completed. Each hook function provides a QueryContext which contains details about the query. Transactions inherit hooks from the database context that started the transaction.
```golang
db.UseHook(
    func(ctx context.Context, qctx ssql.QueryContext, err error) {
        if err != nil {
            fmt.Println(qctx.Function, " query failed")
        } else {
            dur := time.Since(qctx.StartTime)
            fmt.Println(qctx.Function, " query succeeded in ", dur)
        }
    }
)
```