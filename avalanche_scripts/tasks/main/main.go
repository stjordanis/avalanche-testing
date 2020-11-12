package main

import "github.com/ava-labs/avalanche-testing/avalanche_scripts/tasks/locked"

// "github.com/ava-labs/avalanche-testing/avalanche_scripts/tasks/avahub"
// "github.com/ava-labs/avalanche-testing/avalanche_scripts/tasks/subscribe"

func main() {
	locked.SendLockedFunds("/Users/aaronbuchwald/Downloads/phillip_leftover.csv")
	// avahub.SendFunds("/Users/aaronbuchwald/Downloads/avahub_week6.csv")
	// subscribe.DoTask()
}
