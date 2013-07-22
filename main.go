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

TODO

 * Specify host
 * Use Perlin-like noise as values.
 * Randomly skip values
 * Generate same set of metrics each time.

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
var differentMetrics int

func init() {
	flag.IntVar(&minNameParts, "min", 2, "Minimum number of metric name elements")
	flag.IntVar(&maxNameParts, "max", 3, "Maximum number of metric name elements")
	flag.Int64Var(&maxMetrics, "n", 1<<60, "Number of metrics to send")
	flag.IntVar(&differentMetrics, "m", 100, "Number of different metrics to generate")
}

var nameparts = []string{
	"foo",
	"bar",
	"baz",
	"qux",
	"frob",
}

// Simple power function on integers
func intPower(x, y int) int {
	res := 1
	for i := 1; i <= y; i++ {
		res = res * x
	}
	return res
}

// Generate a list of metric names of specified length
func generateMetricNames() []string {
	// How many can we generate witht the given constraints
	num := 0
	for i := minNameParts; i <= maxNameParts; i++ {
		num += intPower(len(nameparts), i)
	}
	if num < differentMetrics {
		fmt.Println("Can only generate", num, "combinations, asked to do", differentMetrics)
		os.Exit(1)
	}

	// Create a map and fill it with random data
	m := make(map[string]bool, differentMetrics)
	for len(m) < differentMetrics {
		m[generateMetricName()] = true
	}

	// Extract keys and return those
	l := []string{}
	for key := range m {
		l = append(l, key)
	}

	return l
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

func main() {
	// Read options
	flag.Parse()
	if minNameParts < 1 {
		minNameParts = 1
	}
	if minNameParts > maxNameParts {
		fmt.Println("Error: -min (%v) should be smaller than -max (%v)", minNameParts, maxNameParts)
		os.Exit(1)
	}

	// Dial up a server
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	index := int64(0)
	timeOffset := int64(0)
	metricNames := generateMetricNames()
	max := maxMetrics

	// Send data
	for index < max {
		for _, metricName := range metricNames {
			// Stop if we've generated enough
			if index >= max {
				break
			}
			metric := fmt.Sprintf(
				"%v %.5v %d\n",
				metricName,
				rand.NormFloat64()*5+5, // Mostly 0 to 10
				time.Now().Unix()+timeOffset,
			)
			fmt.Print(metric)
			fmt.Fprint(conn, metric)
			index++
		}
		timeOffset++
	}
	os.Exit(0)
}
