package main

import (
	"flag"
	"fmt"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/logging"
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
		fmt.Sprintf("Log level to use for the controller (%v)", logging.GetAcceptableStrings()),
	)
	flag.Parse()

	logLevelPtr := logging.LevelFromString(*logLevelArg)
	if logLevelPtr == nil {
		// It's a little goofy that we're logging an error before we've set the loglevel, but we do so at the highest
		//  level so that whatever the default the user should see it
		logrus.Fatal("Invalid initializer log level %v", *logLevelArg)
		os.Exit(1)
	}
	logrus.SetLevel(*logLevelPtr)

	controller := controller.NewTestController(ava_testsuite.AvaTestSuite{})

	logrus.Infof("Running test '%v'...", *testNameArg)
	err := controller.RunTest(*testNameArg, *networkInfoFilepathArg)
	if err != nil {
		logrus.Errorf("Test %v failed:", *testNameArg)
		logrus.Error(err)
		os.Exit(1)
	}
	logrus.Info("Test %v succeeded", *testNameArg)
}
