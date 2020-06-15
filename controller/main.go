package main

import (
	"flag"
	"fmt"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite"
	"github.com/kurtosis-tech/kurtosis/controller"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	TRACE_LOGLEVEL_ARG = "trace"
	DEBUG_LOGLEVEL_ARG = "debug"
	INFO_LOGLEVEL_ARG = "info"
	WARN_LOGLEVEL_ARG = "warn"
	ERROR_LOGLEVEL_ARG = "error"
	FATAL_LOGLEVEL_ARG = "fatal"
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

	logLevelArg := flag.String(
		"log-level",
		"info",
		fmt.Sprintf(
			"Log level to use (%v, %v, %v, %v, %v, %v)",
			TRACE_LOGLEVEL_ARG,
			DEBUG_LOGLEVEL_ARG,
			INFO_LOGLEVEL_ARG,
			WARN_LOGLEVEL_ARG,
			ERROR_LOGLEVEL_ARG,
			FATAL_LOGLEVEL_ARG)
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
	logrus.Info("Test %v succeeded", *testNameArg)
}
