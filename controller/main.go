package main

import (
	"flag"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite"
	"github.com/kurtosis-tech/kurtosis/controller"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	// TODO make this configurable (passed from the initializer)
	logrus.SetLevel(logrus.TraceLevel)

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
	err := controller.RunTest(*testNameArg, *networkInfoFilepathArg)

	if err != nil {
		logrus.Errorf("Test %v failed:", *testNameArg)
		logrus.Error(err)
		os.Exit(1)
	}
	logrus.Infof("Test %v succeeded", *testNameArg)
}
