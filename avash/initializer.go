package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_testsuite/rpc_workflow_test"
	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
)

var (
	errUnhealthy = errors.New("node was not marked as healthy")
)

const (
	stakerClientURI    = "http://127.0.0.1:9660"
	delegatorClientURI = "http://127.0.0.1:9662"
	requestTimeout     = 20 * time.Second
)

func awaitSetup(clients ...*apis.Client) error {
	for _, client := range clients {
		if healthy, _ := client.HealthAPI().AwaitHealthy(4, 3*time.Second); healthy {
			return nil
		}
	}
	return errUnhealthy
}

func main() {
	stakerClient := apis.NewClient(stakerClientURI, requestTimeout)
	delegatorClient := apis.NewClient(delegatorClientURI, requestTimeout)
	if err := awaitSetup(stakerClient, delegatorClient); err != nil {
		fmt.Printf("Test setup failed: %s\n", err)
		return
	}
	rpcWorkflowTest := rpc_workflow_test.NewRPCWorkflowTestExecutor(
		stakerClient,
		delegatorClient,
		20*time.Second,
	)

	if err := rpcWorkflowTest.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
	}
}
