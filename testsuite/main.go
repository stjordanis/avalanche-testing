package main

import (
	"flag"
	"fmt"
	"os"

	testsuite "github.com/ava-labs/avalanche-testing/testsuite/kurtosis"
	"github.com/kurtosis-tech/kurtosis-go/lib/client"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	// --------- Kurtosis-internal params --------------------------------------
	metadataFilepath := flag.String(
		"metadata-filepath",
		"",
		"The filepath of the file in which the test suite metadata should be written")
	testArg := flag.String(
		"test",
		"",
		"The name of the test to run")
	kurtosisAPIIPArg := flag.String(
		"kurtosis-api-ip",
		"",
		"IP address of the Kurtosis API endpoint")
	logLevelArg := flag.String(
		"log-level",
		"",
		"String corresponding to Logrus log level that the test suite will output with",
	)
	servicesDirpathArg := flag.String(
		"services-relative-dirpath",
		"",
		"Dirpath, relative to the root of the suite execution volume, where directories for each service should be created")

	// ----------------------- Avalanche testing-custom params ---------------------------------
	avalancheGoImageArg := flag.String(
		"avalanche-go-image",
		"",
		"Name of Avalanche Go Docker image that will be used to launch Avalanche Go nodes")
	byzantineGoImageArg := flag.String(
		"byzantine-go-image",
		"",
		"Name of Byzantine Avalanche Go Docker image that will be used to launch Avalanche Go nodes with Byzantine behaviour")

	flag.Parse()

	level, err := logrus.ParseLevel(*logLevelArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred parsing the log level string: %v\n", err)
		os.Exit(1)
	}
	logrus.SetLevel(level)

	logrus.Debugf("Byzantine image name: %s", *byzantineGoImageArg)
	testSuite := testsuite.AvalancheTestSuite{
		ByzantineImageName: *byzantineGoImageArg,
		NormalImageName:    *avalancheGoImageArg,
	}
	exitCode := client.Run(testSuite, *metadataFilepath, *servicesDirpathArg, *testArg, *kurtosisAPIIPArg)
	os.Exit(exitCode)
}
