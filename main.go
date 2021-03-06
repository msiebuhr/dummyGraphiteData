/*
Package dummyGraphiteData is a program that sends fake graphite data to localhost:2003.

Usage:

	dummyGraphiteData [-min=2] [-max=3] [-n=Inf]

Where

	-min is the minimum elements in the metric name
	-max is the maximum elements in the metric name
	-n is the number of metrics to send

Currently noise is generated by Perlin1d(unix timestamp + FNV(metric name)) + FNV(metric name).

TODO

- Specify host
- Randomly skip values
- Generate same set of metrics each time.

*/
package main

import (
	"flag"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
	"io"
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
		fmt.Printf("Error: -min (%v) should be smaller than -max (%v)\n", minNameParts, maxNameParts)
		os.Exit(1)
	}

	// Dial up a server
	conn, err := net.Dial("tcp", "localhost:2003")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

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
			hash := fnv.New32()
			hash.Write([]byte(metricName))
			//fmt.Println(int64(hash.Sum32()) + timeOffset)
			perlinIndex := float64(int64(hash.Sum32()) + timeOffset)
			metric := fmt.Sprintf(
				"%v %.5v %d\n",
				metricName,
				PerlinNoise_1D(perlinIndex)+float64((int64(hash.Sum32())%10)-5),
				time.Now().Unix()+timeOffset,
			)
			fmt.Print(metric)
			fmt.Fprint(conn, metric)
			index++
		}

		// Try reading a value from the socket to see if it's closed
		conn.SetReadDeadline(time.Now().Add(time.Second))
		_, err := conn.Read([]byte{})
		if err != nil && err != io.EOF {
			fmt.Println("Network error:", err.Error())
			break;
		}

		timeOffset++
	}
}
