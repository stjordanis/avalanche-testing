package main

import (
	"flag"
	// "github.com/kurtosis-tech/kurtosis/ava_commons/testsuite"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	testNameArg := flag.String(
		"test",
		"",
		"Comma-separated list of specific tests to run (leave empty or omit to run all tests)",
	)

	networkInfoFilepathArg := flag.String(
		"network-info-filepath",
		"",
		"Filepath of file containing JSON-serialized representation of the network of service Docker containers",
	)
	flag.Parse()

	if _, err := os.Stat(*networkInfoFilepathArg); err != nil {
		panic("Nonexistent file: " + *networkInfoFilepathArg)
	}

	println(fmt.Sprintf("Would run %v:" , *testNameArg))

	data, err := ioutil.ReadFile(*networkInfoFilepathArg)
	if err != nil {
		// TODO make this a proper error
		panic("Could not read file bytes!")
	}
	println(fmt.Sprintf("Contents of file: %v", string(data)))
}
