package main

import (
	"encoding/gob"
	"flag"
	"github.com/kurtosis-tech/kurtosis/ava_commons/testsuite"
	"github.com/kurtosis-tech/kurtosis/commons/testnet"
	testsuite2 "github.com/kurtosis-tech/kurtosis/commons/testsuite"
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

	fp, err := os.Open(*networkInfoFilepathArg)
	if err != nil {
		panic("Could not open network info file for reading")
	}
	decoder := gob.NewDecoder(fp)

	var rawServiceNetwork testnet.RawServiceNetwork
	err = decoder.Decode(&rawServiceNetwork)
	if err != nil {
		panic("Decoding raw service network information failed")
	}

	testConfig, found := testsuite.AvaTestSuite{}.GetTests()[*testNameArg]
	if !found {
		panic("Nonexistent test: " + *testNameArg)
	}

	untypedNetwork, err := testConfig.NetworkLoader.LoadNetwork(rawServiceNetwork.ServiceIPs)
	if err != nil {
		panic("Unable to load network from service IPs")
	}

	testSucceeded := true
	context := testsuite2.TestContext{}
	testConfig.Test.Run(untypedNetwork, context)
	defer func() {
		if result := recover(); result != nil {
			testSucceeded = false
		}
	}()

	if !testSucceeded {
		os.Exit(1)
	}
}
