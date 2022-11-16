package drape

import (
	"context"
	"time"
)

type QueryHook func(context.Context, QueryDetails, error)

type QueryDetails struct {
	StartTime time.Time     // start time of the query
	Query     string        // query string
	Params    []interface{} // parameters of the query
}

