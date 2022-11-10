package cli

import (
	"context"

	"gfx.cafe/open/ghost"
)

type CmdCtx struct {
	c   ghost.Client
	ans any

	context.Context
}

func (c *CmdCtx) Return(ans any) {
	c.ans = ans
}

func Ans[T any](c *CmdCtx) T {
	if res, ok := c.ans.(T); ok {
		return res
	}
	return *new(T)
}
