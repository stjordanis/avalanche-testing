package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-testing/avalanche_client/apis"
	"github.com/ava-labs/avalanche-testing/testsuite/tests/bombard"
)

var (
	errUnhealthy = errors.New("node was not marked as healthy")
)

const (
	stakerClientURI    = "http://127.0.0.1:9660"
	delegatorClientURI = "http://127.0.0.1:9662"
	requestTimeout     = 10 * time.Second
)

// cd $GOPATH
// go get -v -d github.com/ava-labs/gecko (if you do not already have gecko installed)
// go get -v -d github.com/ava-labs/avalanche-testing
// cd $GOPATH/src/github.com/ava-labs/gecko
// git checkout master (from public repo)
// cd $GOPATH/src/github.com/ava-labs/avalanche-testing
// go mod download
// copy the contents of this file into a file called bombard.go
// go run bombard.go

// wait until each node reports healthy and return an error if it takes too long
func awaitSetup(clients ...*apis.Client) error {
	for _, client := range clients {
		if healthy, _ := client.HealthAPI().AwaitHealthy(4, 3*time.Second); healthy {
			return nil
		}
	}
	return errUnhealthy
}

// return clients to interact with each avash node
func fiveNodeAvashClients() []*apis.Client {
	clients := make([]*apis.Client, 5)
	for i := 0; i < 5; i++ {
		uri := fmt.Sprintf("http://127.0.0.1:%d", 9650+i*2)
		clients[i] = apis.NewClient(uri, requestTimeout)
	}
	return clients
}

// return clients for each of the devnet nodes
func devNetClients() []*apis.Client {
	return []*apis.Client{
		apis.NewClient("http://3.95.247.197:21000", requestTimeout), // Commented out because this node is not responding to requests
		apis.NewClient("http://35.153.99.244:21000", requestTimeout),
		apis.NewClient("http://34.201.137.119:21000", requestTimeout),
		apis.NewClient("http://54.146.1.110:21000", requestTimeout),
		apis.NewClient("http://54.91.255.231:21000", requestTimeout),
	}
}

func bombardNetwork(clients []*apis.Client, numTxs, txFee uint64, done chan error) {
	executor := bombard.NewBombardExecutor(clients, numTxs, txFee, 10*time.Second)
	if err := executor.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
		done <- fmt.Errorf("bombardNetwork failed due to: %s", err)
		return
	}
	done <- nil
}

func main() {
	// clients := fiveNodeAvashClients()
	clients := devNetClients()

	if err := awaitSetup(clients...); err != nil {
		fmt.Printf("Test did not run due to: %s\n", err)
		return
	}

	executor := bombard.NewBombardExecutor(clients, 100000, 1000000, 15*time.Second)
	if err := executor.ExecuteTest(); err != nil {
		fmt.Printf("Test failed: %s\n", err)
	}
}
