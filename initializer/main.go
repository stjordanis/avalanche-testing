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
		*testControllerImageNameArg,)

	// Create the container based on the configurations, but don't start it yet.
	fmt.Println("I'm going to run a Gecko testnet, and hang while it's running! Kill me and then clear your docker containers.")
	results, error := testSuiteRunner.RunTests(testNames)
	if error != nil {
		panic(error)
	}

	logrus.Info("================================== TEST RESULTS ================================")
	allTestsSucceeded := true
	for testName, result := range results {
		logrus.Infof("- %v: %v", testName, result)
		allTestsSucceeded = allTestsSucceeded && result == initializer.PASSED
	}

	if allTestsSucceeded {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
