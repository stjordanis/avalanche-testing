package main

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche/services"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/cchain"
)

func main() {
	ipAddr := "127.0.0.1"
	port := 9650
	requestTimeout := 3 * time.Second

	client, err := services.NewClient(ipAddr, port, requestTimeout)
	if err != nil {
		fmt.Printf("failed to create client: %s\n", err)
		return
	}

	ethAPI := client.CChainEthAPI()
	test1 := cchain.NewEthAPIExecutor(ethAPI)
	if err := test1.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
		return
	}

	test2 := cchain.CreateAtomicWorkflowTest(client, 1000000)
	if err := test2.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
		return
	}

	test3 := cchain.NewBasicTransactionThroughputTest(client, 3, 200)
	if err := test3.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
		return
	}
}
