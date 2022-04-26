package client

import (
	"context"
	"elastiq/config"
	"elastiq/query"
	"io"
)

type Client interface {
	Query(ctx context.Context, env *config.Env, q *query.Query, o query.Options) (io.Reader, error)
}
