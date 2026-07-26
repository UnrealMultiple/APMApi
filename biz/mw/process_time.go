package mw

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// ProcessTime 与老API保持一致的 X-Process-Time 响应头
func ProcessTime() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		c.Next(ctx)
		ms := float64(time.Since(start).Nanoseconds()) / 1e6
		c.Header("X-Process-Time", strconv.FormatFloat(ms, 'f', -1, 64)+"ms")
	}
}
