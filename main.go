/*
Package dummyGraphiteData is a program that sends fake graphite data to localhost:8000.

Usage:

	dummyGraphiteData [-min=2] [-max=3] [-n=Inf]

Where

    -min is the minimum elements in the metric name
	-max is the maximum elements in the metric name
	-n is the number of metrics to send

Currently, it generates data in a very crappy way (generates a fake metric
name, random value and the previosu metrics' timestamp bumped by one second)

TODO: User should supply number of different metrics to send
TODO: Loop over generated metric names, use Perlin-like noise as values and
      randomly skip entries.
*/
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

var minNameParts int
var maxNameParts int
var maxMetrics int64

func init() {
	flag.IntVar(&minNameParts, "min", 2, "Minimum number of metric name elements")
	flag.IntVar(&maxNameParts, "max", 3, "Maximum number of metric name elements")
	flag.Int64Var(&maxMetrics, "n", 1<<60, "Number of metrics to send")
}

var nameparts = []string{
	"foo",
	"bar",
	"baz",
	"qux",
}

func generateMetricName() string {
	// Length of mame
	length := maxNameParts + rand.Intn(maxNameParts-minNameParts+1)

	parts := make([]string, length)
	for i := 0; i < length; i++ {
		parts[i] = nameparts[rand.Intn(len(nameparts))]
	}

	return strings.Join(parts, ".")
}

func generateMetric(timeOffset int64) string {
	return fmt.Sprintf(
		"%v %.5v %d",
		generateMetricName(),
		rand.NormFloat64()*5+5, // Mostly 0 to 10
		time.Now().Unix()+timeOffset,
	)
}

func main() {
	// Read options
	flag.Parse()
	if minNameParts < 1 {
		minNameParts = 1
	}
	if minNameParts > maxNameParts {
		fmt.Println("Error: -min (%v) should be smaller than -max (%v)", minNameParts, maxNameParts)
	}

	// Dial up a server
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	index := int64(0)
	max := maxMetrics

	// Send data
	for index < max {
		metric := generateMetric(index)
		fmt.Println(metric)
		fmt.Fprintf(conn, metric)
		fmt.Fprintf(conn, "\n")
		index++
	}
	os.Exit(0)
}
