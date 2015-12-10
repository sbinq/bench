package bench

import (
	"fmt"
	"time"
	"bytes"
	
	"github.com/codahale/hdrhistogram"
)

// Context for each benchmarking run
type Context struct {
	ID int
	Iteration int64
	histogram *hdrhistogram.Histogram
}

func newContext(id int) *Context {
	t := Context{ID:id}
	t.histogram = hdrhistogram.New(min, max, resolution)
	return &t
}

// RecordTime records the time taken for one benchmarking run.
func (c *Context) RecordTime(t time.Duration) {
	c.histogram.RecordValue(t.Nanoseconds())
}

func (c *Context) String() string {
	prefix := "  "
	var buf bytes.Buffer
	percentiles := []float64{50, 99.9, 100}

	fmt.Fprintf(&buf, "Task: %d, Total runs: %d\n", c.ID, c.Iteration)
	for _, p := range percentiles {
		fmt.Fprintf(&buf, "%s%2.1fth percentile: %.2fms\n", prefix, p, float32(c.histogram.ValueAtQuantile(p))/1000000.0)
	}
	return buf.String()
}
