package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/tests"
	"github.com/ava-labs/avalanche-e2e-tests/commons/logging"
	"github.com/kurtosis-tech/kurtosis/initializer"
	"github.com/sirupsen/logrus"
)

const (
	testNameArgSeparator     = ","
	geckoImageNameEnvVar     = "GECKO_IMAGE_NAME"
	byzantineImageNameEnvVar = "BYZANTINE_IMAGE_NAME"
	defaultParallelism       = 4

	// The number of bits to make each test network, which dictates the max number of services a test can spin up
	// Here we choose 8 bits = 256 max services per test
	networkWidthBits = 8
)

/*
A CLI intended to be the main entrypoint into running the Avalanche E2E test suite.
*/
func main() {
	// NOTE: we'll need to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	doListArg := flag.Bool(
		"list",
		false,
		"Rather than running the tests, lists the tests available to run",
	)

	// Define and parse command line flags.
	geckoImageNameArg := flag.String(
		"gecko-image-name",
		"",
		"The name of a pre-built Gecko image, either on the local Docker engine or in Docker Hub",
	)

	byzantineImageNameArg := flag.String(
		"byzantine-image-name",
		"",
		"The name of a pre-built Byzantine Gecko image, on the local Docker engine",
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

	logrus.Info("Welcome to the Avalanche E2E test suite, powered by the Kurtosis framework")
	testSuite := tests.AvaTestSuite{
		ByzantineImageName: *byzantineImageNameArg,
		NormalImageName:    *geckoImageNameArg,
	}
	if *doListArg {
		testNames := []string{}
		for name, _ := range testSuite.GetTests() {
			testNames = append(testNames, name)
		}
		sort.Strings(testNames)

		for _, name := range testNames {
			fmt.Println("- " + name)
		}
		os.Exit(0)
	}

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

	testNamesArgStr := strings.TrimSpace(*testNamesArg)
	testNames := map[string]bool{}
	if len(testNamesArgStr) > 0 {
		testNamesList := strings.Split(testNamesArgStr, testNameArgSeparator)
		for _, name := range testNamesList {
			testNames[name] = true
		}
	}

	testSuiteRunner := initializer.NewTestSuiteRunner(
		testSuite,
		*testControllerImageNameArg,
		*controllerLogLevelArg,
		map[string]string{
			geckoImageNameEnvVar:     *geckoImageNameArg,
			byzantineImageNameEnvVar: *byzantineImageNameArg,
		},
		networkWidthBits)

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
