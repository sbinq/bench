//Package bench makes it easy to write benchamrking apps.
package bench

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/codahale/hdrhistogram"
	"golang.org/x/net/context"
)

// Bench holds data to be used for benchmarking. It can perform aggresively wherein it repeatedly sends calls the benchmarking function, or it can perform a uniform bench wherein it can invoke the benchmarking function at a uniform rate.
type Bench struct {
	// values configured by the user
	concurrentRuns int
	duration       time.Duration
	toBenchmark    func(t *Context)

	// context for individual runs
	runContexts []*Context

	// aggregated metrics
	timers   map[string]*hdrhistogram.Histogram
	counters map[string]int64

	// total calls made to each task
	calls int64
	// time taken for the full run
	timeTaken time.Duration
}

// NewBench creates a new instance of Bench
func NewBench(concurrency int, duration time.Duration, toBenchmark func(*Context)) *Bench {
	b := &Bench{
		concurrentRuns: concurrency,
		duration:       duration,
		toBenchmark:    toBenchmark,
	}

	b.timers = make(map[string]*hdrhistogram.Histogram)
	b.counters = make(map[string]int64)

	for i := 0; i < b.concurrentRuns; i++ {
		b.runContexts = append(b.runContexts, newContext(i+1))
	}

	return b
}

// Run the benchmark
func (b *Bench) Run() {

	// Only run it once
	if b.calls != 0 {
		return
	}

	var wg sync.WaitGroup
	start := time.Now()

	ctx, _ := context.WithTimeout(context.Background(), b.duration)

	for i := 1; i <= b.concurrentRuns; i++ {
		go func(ctx context.Context, runContext *Context) {
			wg.Add(1)
			defer wg.Done()

			for j := 1; ; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					runContext.Iteration = int64(j)
					b.toBenchmark(runContext)
				}
			}
		}(ctx, b.runContexts[i-1])
	}

	wg.Wait()
	b.timeTaken = time.Since(start)

	b.aggregate()
}

// UniformRun calls the function at the specified rate
func (b *Bench) UniformRun(rps int) {

	// Only run it once
	if b.calls != 0 {
		return
	}

	var wg sync.WaitGroup
	tokens := make(chan bool, 10*rps)
	defer close(tokens)

	ctx, _ := context.WithTimeout(context.Background(), b.duration)
	ticker := time.NewTicker(time.Second / time.Duration(rps))
	defer ticker.Stop()

	start := time.Now()

	// generate a queue of events for consumption by the runs
	go func(ctx context.Context, ticker *time.Ticker) {
		wg.Add(1)
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tokens <- true
			}
		}
	}(ctx, ticker)

	// start the runs
	for i := 1; i <= b.concurrentRuns; i++ {
		wg.Add(1)
		go func(ctx context.Context, runContext *Context) {
			defer wg.Done()

			for j := 1; ; j++ {
				select {
				case <-ctx.Done():
					return
				case <-tokens:
					runContext.Iteration = int64(j)
					b.toBenchmark(runContext)
				}
			}
		}(ctx, b.runContexts[i-1])
	}

	wg.Wait()
	b.timeTaken = time.Since(start)

	b.aggregate()
}

// aggregate run contexts
func (b *Bench) aggregate() {

	// aggregate timer metrics
	for n := range b.runContexts[0].timers {
		t := hdrhistogram.New(min, max, resolution)
		b.timers[n] = t
		for i := 0; i < b.concurrentRuns; i++ {
			t.Merge(b.runContexts[i].timers[n])
		}
	}

	// aggregate counters
	for n := range b.runContexts[0].counters {
		for i := 0; i < b.concurrentRuns; i++ {
			b.counters[n] += b.runContexts[i].counters[n]
		}
	}

	// aggregate call counts
	for i := 0; i < b.concurrentRuns; i++ {
		b.calls += b.runContexts[i].Iteration
	}
}

// String converts the output of the bench into a printable form
func (b *Bench) String() string {
	prefix := "  "
	var buf bytes.Buffer
	percentiles := []float64{50, 99.9, 100}

	fmt.Fprintf(&buf, "Duration: %2.2fs, Concurrency: %d, Total runs: %d\n", b.timeTaken.Seconds(), b.concurrentRuns, b.calls)

	for n, h := range b.timers {
		fmt.Fprintf(&buf, "%s>>Timer: %s \n", prefix, n)
		for _, p := range percentiles {
			fmt.Fprintf(&buf, "%s%s%2.1fth percentile: %.2fms\n", prefix, prefix, p, float32(h.ValueAtQuantile(p))/1000000.0)
		}
	}
	for n, count := range b.counters {
		fmt.Fprintf(&buf, "%s>>Counter: %s\n", prefix, n)
		fmt.Fprintf(&buf, "%s%sValue: %d \n", prefix, prefix, count)
	}
	return buf.String()
}
