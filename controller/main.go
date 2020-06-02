package main

import (
	"flag"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite"
	"github.com/kurtosis-tech/kurtosis/controller"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

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

	logrus.Infof("Running test '%v'...", *testNameArg)

	controller := controller.NewTestController(ava_testsuite.AvaTestSuite{})
	// TODO replace the boolean result with detailed information about the test suite results
	succeeded, err := controller.RunTests(*testNameArg, *networkInfoFilepathArg)
	if err != nil {
		panic(err)
	}

	if !succeeded {
		os.Exit(1)
	}
}
