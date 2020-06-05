package main

import (
	"flag"
	"fmt"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_testsuite"
	"github.com/kurtosis-tech/kurtosis/initializer"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)


const (
	DEFAULT_STARTING_PORT = 9650
	DEFAULT_ENDING_PORT = 10650

	TEST_NAME_ARG_SEPARATOR = ","
)

func main() {
	// TODO make this configurable
	logrus.SetLevel(logrus.TraceLevel)

	fmt.Println("Welcome to Kurtosis E2E Testing for Ava.")

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
	portRangeStartArg := flag.Int(
		"port-range-start",
		DEFAULT_STARTING_PORT,
		"Beginning of port range to be used by testnet on the local environment. Must be between 1024-65535",
	)

	portRangeEndArg := flag.Int(
		"port-range-end",
		DEFAULT_ENDING_PORT,
		"End of port range to be used by testnet on the local environment. Must be between 1024-65535",
	)

	testNamesArg := flag.String(
		"test-names",
		"",
		"Comma-separated list of test names to run (default or empty: run all tests)",
	)
	flag.Parse()

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
		*portRangeStartArg,
		*portRangeEndArg)

	// Create the container based on the configurations, but don't start it yet.
	fmt.Println("I'm going to run a Gecko testnet, and hang while it's running! Kill me and then clear your docker containers.")
	results, error := testSuiteRunner.RunTests(testNames)
	if error != nil {
		panic(error)
	}

	logrus.Info("=========== TEST RESULTS ============")
	allTestsSucceeded := true
	for testName, result := range results {
		// TODO get information about why stuff failed
		logrus.Infof("- %v: %v", testName, result)
		allTestsSucceeded = allTestsSucceeded && result == initializer.PASSED
	}

	if allTestsSucceeded {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
