package main

import (
	"flag"
	"fmt"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/logging"
	"github.com/kurtosis-tech/kurtosis/initializer"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)


const (
	TEST_NAME_ARG_SEPARATOR = ","

	defaultParallelism = 4
)

func main() {
	// NOTE: we'll need to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		FullTimestamp:             true,
	})

	// Define and parse command line flags.
	geckoImageNameArg := flag.String(
		"gecko-image-name", 
		"",
		"The name of a pre-built Gecko image, either on the local Docker engine or in Docker Hub",
	)

	testControllerImageNameArg := flag.String(
		"test-controller-image-name",
		"",
		"The name of a pre-built test controller image, either on the local Docker engine or in Docker Hub",
	)

	testNamesArg := flag.String(
		"test-names",
		"",
		"Comma-separated list of test names to run (default or empty: run all tests)",
	)

	initializerLogLevelArg := flag.String(
		"initializer-log-level",
		"debug",
		fmt.Sprintf("Log level to use for the initializer (%v)", logging.GetAcceptableStrings()),
	)

	controllerLogLevelArg := flag.String(
		"controller-log-level",
		"debug",
		fmt.Sprintf("Log level to use for the initializer (%v)", logging.GetAcceptableStrings()),
	)

	parallelismArg := flag.Uint(
		"parallelism",
		defaultParallelism,
		"Number of tests to run in parallel",
	)

	flag.Parse()

	initializerLevelPtr := logging.LevelFromString(*initializerLogLevelArg)
	if initializerLevelPtr == nil {
		// It's a little goofy that we're logging an error before we've set the loglevel, but we do so at the highest
		//  level so that whatever the default the user should see it
		logrus.Fatalf("Invalid initializer log level %v", *initializerLogLevelArg)
		os.Exit(1)
	}
	logrus.SetLevel(*initializerLevelPtr)

	// Technically this validation should be done only in the controller (the initializer shouldn't know anything about
	//  what logging the controller uses) but we do this here to save the user from needing to wait for a controller to
	//  start up to find out they typo'd the log level
	controllerLevelPtr := logging.LevelFromString(*controllerLogLevelArg)
	if controllerLevelPtr == nil {
		logrus.Fatalf("Invalid controller log level %v", *controllerLogLevelArg)
		os.Exit(1)
	}


	logrus.Info("Welcome to the Ava E2E test suite, powered by the Kurtosis framework")
	testNamesArgStr := strings.TrimSpace(*testNamesArg)
	var testNames []string
	if len(testNamesArgStr) == 0 {
		testNames = make([]string, 0, 0)
	} else {
		testNames = strings.Split(testNamesArgStr, TEST_NAME_ARG_SEPARATOR)
	}

	testSuiteRunner := initializer.NewTestSuiteRunner(
		ava_testsuite.AvaTestSuite{},
		*geckoImageNameArg,
		*testControllerImageNameArg,
		*controllerLogLevelArg)

	// Create the container based on the configurations, but don't start it yet.
	allTestsSucceeded, error := testSuiteRunner.RunTests(testNames, *parallelismArg)
	if error != nil {
		logrus.Error("An error occurred running the tests:")
		logrus.Error(error)
		os.Exit(1)
	}

	if allTestsSucceeded {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
