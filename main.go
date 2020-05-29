package main

import (
	// "github.com/kurtosis-tech/kurtosis/ava_commons/testsuite"
	"fmt"
	"os"
)

func main() {
	testName := os.Args[1]
    // testSuite := testsuite.AvaTestSuite{}
    println(fmt.Sprintf("Would run %v", testName))
}
