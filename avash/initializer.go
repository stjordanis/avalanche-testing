package main

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/rpc_workflow_test"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
)

const (
	stakerClientURI    = "http://127.0.0.1:9660"
	delegatorClientURI = "http://127.0.0.1:9662"
	requestTimeout     = 10 * time.Second
)

func main() {
	stakerClient := apis.NewClient(stakerClientURI, requestTimeout)
	delegatorClient := apis.NewClient(delegatorClientURI, requestTimeout)
	rpcWorkflowTest := rpc_workflow_test.NewRPCWorkflowTestExecutor(
		stakerClient,
		delegatorClient,
		20*time.Second,
	)

	if err := rpcWorkflowTest.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
	}
}
