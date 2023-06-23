package hold

import (
	"sync/atomic"
)

type Counter struct {
	count int64
}

func (c *Counter) Get() int64 {
	return atomic.LoadInt64(&c.count)
}

func (c *Counter) Add(count int64) int64 {
	return atomic.AddInt64(&c.count, int64(count))
}
