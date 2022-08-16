package proxy

import (
	"context"
	"time"
)

type Context struct {
	ctx context.Context
}

func (c Context) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (c Context) Done() <-chan struct{} {
	return nil
}

func (c Context) Err() error {
	return nil
}

func (c Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func WithoutCancel(ctx context.Context) context.Context {
	return Context{ctx: ctx}
}
