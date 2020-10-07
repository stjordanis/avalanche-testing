package main

import (
	// "github.com/ava-labs/avalanche-testing/avalanche_scripts/tasks/locked"
	// "github.com/ava-labs/avalanche-testing/avalanche_scripts/tasks/avahub"
	"github.com/ava-labs/avalanche-testing/avalanche_scripts/tasks/subscribe"
)

func main() {
	// locked.SendLockedFunds("/Users/aaronbuchwald/Documents/go/src/github.com/ava-labs/avalanche-testing/avalanche_scripts/tasks/locked/amounts.csv")
	// avahub.SendFunds("/Users/aaronbuchwald/Downloads/avahub.csv")
	subscribe.DoTask()
}
