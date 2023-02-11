package ssql

import (
	"context"
	"time"
)

type QueryHook func(context.Context, QueryContext, error)

type QueryContext struct {
	StartTime time.Time     // start time of the query
	Function  string        // function called
	Query     string        // query string
	Params    []interface{} // parameters of the query
}
