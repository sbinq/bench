
# Bench [![Build Status](https://travis-ci.org/chirayu/bench.svg?branch=master)](https://travis-ci.org/chirayu/bench) [![GoDoc](http://godoc.org/github.com/chirayu/bench?status.png)](http://godoc.org/github.com/chirayu/bench)

**Bench** is a golang package to help speed up benchmarking projects. You can use it to do a quick benchmark of your database, queue or http server.

## Features

Bench repeatedly calls a function with the specificied concurrency. For each call, it captures metrics, which it then summarizes at the end of the run. Metrics captured include:

* latency for percentiles. By default it displays 50th, 90th and 99th percentiles. 
* calls per second 

Bench provides a Context for each function call, which contains a counter that is incremented on each call. Use it to

* Pick different between urls for each call
* Add it as a custom header for your Rest call
* Use it to generate different DB data

## Usage

Bench is easy to integrate with your code. This code benchmarks a http server with a concurrency of 10 for 10 seconds.

```golang
func benchmarkMe(c *bench.Context) {
	start := time.Now()
	resp, err := http.Get("http://www.flipkart.com")
	if err != nil {
		fmt.Printf("Error %s", err)
	}
	defer resp.Body.Close()
	c.RecordTime(time.Since(start))
}

func main() {
	b := bench.NewBench(10, time.Second*10, benchmarkMe)
	b.Run()
	fmt.Printf("%s", b)
```

## Upcoming features

Two features are in the offing.

1. Ability to capture multiple metrics during runtime. Apart from latency, a http client may want to record data size and error codes. 
2. Make requests at a uniform rate. This is useful when you want to see if you server can sustain a load of X per second. 