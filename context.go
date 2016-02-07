package bench

import (
	"bytes"
	"fmt"
	"time"

	"github.com/codahale/hdrhistogram"
)

// Context for each benchmarking run
type Context struct {
	ID        int
	Iteration int64
	timers    map[string]*hdrhistogram.Histogram
	counters  map[string]int64

	// user values
	Values map[string]interface{}
}

func newContext(id int) *Context {
	c := Context{ID: id}
	c.timers = make(map[string]*hdrhistogram.Histogram)
	c.counters = make(map[string]int64)
	c.Values = make(map[string]interface{})
	return &c
}

// Timer records the time for the specified event
func (c *Context) Timer(name string, t time.Duration) {
	if _, ok := c.timers[name]; ok == false {
		c.timers[name] = hdrhistogram.New(min, max, precision)
	}
	c.timers[name].RecordValue(t.Nanoseconds())
}

// Incr increments a counter
func (c *Context) Incr(name string, v int64) {
	c.counters[name] += v
}

func (c *Context) String() string {
	prefix := "  "
	var buf bytes.Buffer
	percentiles := []float64{50, 99.9, 100}

	fmt.Fprintf(&buf, "Task: %d, Total runs: %d\n", c.ID, c.Iteration)
	for n, h := range c.timers {
		fmt.Fprintf(&buf, "%sTimer: %s \n", prefix, n)
		for _, p := range percentiles {
			fmt.Fprintf(&buf, "%s%2.1fth percentile: %.2fms\n", prefix, p, float64(h.ValueAtQuantile(p))/1000000.0)
		}
	}
	for n, count := range c.counters {
		fmt.Fprintf(&buf, "%sCounter: %s, value: %d \n", prefix, n, count)
	}
	return buf.String()
}

func timerKeys(c *Context) []string {
	ks := make([]string, 0)
	for k := range c.timers {
		ks = append(ks, k)
	}
	return ks
}

func counterKeys(c *Context) []string {
	ks := make([]string, 0)
	for k := range c.counters {
		ks = append(ks, k)
	}
	return ks
}
