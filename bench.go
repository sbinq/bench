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

	// histogram for the benchmarking run
	histogram *hdrhistogram.Histogram
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
		histogram:      hdrhistogram.New(min, max, resolution),
	}

	for i := 0; i < b.concurrentRuns; i++ {
		b.runContexts = append(b.runContexts, newContext(i+1))
	}

	return b
}

// Run the benchmark
func (b *Bench) Run() {
	var wg sync.WaitGroup
	start := time.Now()

	ctx, _ := context.WithTimeout(context.Background(), b.duration)

	for i := 1; i <= b.concurrentRuns; i++ {
		wg.Add(1)
		go func(ctx context.Context, runContext *Context) {
			defer wg.Done()

			for j := 1; ; j++ {
				runContext.Iteration = int64(j)
				select {
				case <-ctx.Done():
					return
				default:
					b.toBenchmark(runContext)
				}
			}
		}(ctx, b.runContexts[i-1])
	}

	wg.Wait()

	// merge task specific data
	for i := 0; i < b.concurrentRuns; i++ {
		b.histogram.Merge(b.runContexts[i].histogram)
		b.calls += b.runContexts[i].Iteration
	}

	b.timeTaken = time.Since(start)
}

// String converts the output of the bench into a printable form
func (b *Bench) String() string {
	prefix := "  "
	var buf bytes.Buffer
	percentiles := []float64{50, 99.9, 100}

	fmt.Fprintf(&buf, "Duration: %2.2fs, Concurrency: %d, Total runs: %d\n", b.timeTaken.Seconds(), b.concurrentRuns, b.calls)
	for _, p := range percentiles {
		fmt.Fprintf(&buf, "%s%2.1fth percentile: %.2fms\n", prefix, p, float32(b.histogram.ValueAtQuantile(p))/1000000.0)
	}
	fmt.Fprintf(&buf, "%s%d calls in %.2fs\n", prefix, b.calls, b.timeTaken.Seconds())
	return buf.String()
}
