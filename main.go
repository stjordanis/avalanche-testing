package main

import (
	// "github.com/kurtosis-tech/kurtosis/ava_commons/testsuite"
	"fmt"
	"os"
)

func main() {
	testName := os.Args[1]
	networkInfoFilepath := os.Args[2]
    println(fmt.Sprintf("Would run %v using network data filepath %v", testName, networkInfoFilepath))
}
