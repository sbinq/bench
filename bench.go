//Package bench makes it easy to write benchamrking apps.
package bench

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/codahale/hdrhistogram"
	"golang.org/x/net/context"
)

// benchmarking function
type benchmarkFn func(t *Context)

// Bench holds data to be used for benchmarking.
// It can perform aggresively wherein it repeatedly sends calls the benchmarking function, or it can perform a uniform bench wherein it can invoke the benchmarking function at a uniform rate.
type Bench struct {
	// values configured by the user
	concurrentRuns int
	duration       time.Duration
	fn             benchmarkFn
	rps            int

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
func NewBench(concurrency int, duration time.Duration, rps int, fn benchmarkFn) *Bench {
	b := &Bench{
		concurrentRuns: concurrency,
		duration:       duration,
		fn:             fn,
		rps:            rps,
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
	var ts *TokenStream
	ctx, _ := context.WithTimeout(context.Background(), b.duration)

	if b.rps > 0 {
		ts = NewTokenStream(b.rps)
		defer ts.Stop()
	}

	// start the runs
	start := time.Now()
	for i := 1; i <= b.concurrentRuns; i++ {
		if b.rps > 0 {
			wg.Add(1)
			go onToken(ctx, b.runContexts[i-1], &wg, ts, b.fn)
		} else {
			wg.Add(1)
			go continuous(ctx, b.runContexts[i-1], &wg, b.fn)
		}
	}

	wg.Wait()
	b.timeTaken = time.Since(start)

	b.aggregate()
}

// aggregate run contexts
func (b *Bench) aggregate() {

	// aggregate timer metrics
	for _, n := range b.allContextsKeys(timerKeys) {
		t := hdrhistogram.New(min, max, precision)
		b.timers[n] = t
		for i := 0; i < b.concurrentRuns; i++ {
			otherTimer, ok := b.runContexts[i].timers[n]
			if ok {
				t.Merge(otherTimer)
			}
		}
	}

	// aggregate counters
	for _, n := range b.allContextsKeys(counterKeys) {
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
	percentiles := []float64{5, 50, 70, 90, 95, 99, 99.9, 99.95, 99.99, 100}

	if b.rps <= 0 {
		fmt.Fprintf(&buf, "Duration: %2.2fs, Concurrency: %d, Total runs: %d\n", b.timeTaken.Seconds(), b.concurrentRuns, b.calls)
	} else {
		fmt.Fprintf(&buf, "Rate: %d calls/sec, Duration: %2.2fs, Concurrency: %d, Total runs: %d\n", b.rps, b.timeTaken.Seconds(), b.concurrentRuns, b.calls)
	}

	for n, h := range b.timers {
		fmt.Fprintf(&buf, "%s>>Timer: %s \n", prefix, n)
		for _, p := range percentiles {
			fmt.Fprintf(&buf, "%s%s%2.2fth percentile: %.2fms\n", prefix, prefix, p, float64(h.ValueAtQuantile(p))/1000000)
		}
		fmt.Fprintf(&buf, "%s%sMean: %.2fms\n", prefix, prefix, float64(h.Mean())/1000000.0)
	}
	for n, count := range b.counters {
		fmt.Fprintf(&buf, "%s>>Counter: %s\n", prefix, n)
		fmt.Fprintf(&buf, "%s%sValue: %d \n", prefix, prefix, count)
	}
	return buf.String()
}

// continuous invokes the benchmarking function continously
// unlike onToken, it doesn't read from a token stream to minimise the overhead of creating and populating a channel
func continuous(ctx context.Context, runContext *Context, wg *sync.WaitGroup, fn benchmarkFn) {
	defer wg.Done()

	for j := 1; ; j++ {
		select {
		case <-ctx.Done():
			return
		default:
			runContext.Iteration = int64(j)
			fn(runContext)
		}
	}
}

// onToken invokes the benchmarking function whenever it gets a token
func onToken(ctx context.Context, runContext *Context, wg *sync.WaitGroup, ts *TokenStream, fn benchmarkFn) {
	defer wg.Done()
	for j := 1; ; j++ {
		select {
		case <-ctx.Done():
			return
		case <-ts.S:
			runContext.Iteration = int64(j)
			fn(runContext)
		}
	}
}

func (b *Bench) allContextsKeys(getCtxKeys func(c *Context) []string) []string {
	m := make(map[string]struct{})
	for _, c := range b.runContexts {
		for _, k := range getCtxKeys(c) {
			m[k] = struct{}{}
		}
	}

	ks := make([]string, 0)
	for k := range m {
		ks = append(ks, k)
	}

	sort.Strings(ks)
	return ks
}
