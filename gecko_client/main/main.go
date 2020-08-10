package main

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis/admin"
)

func main() {
	uri := "http://127.0.0.1:9650"
	timeout := 2 * time.Second

	admin := admin.NewClient(uri, timeout)
	success, err := admin.StartCPUProfiler()
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		return
	}

	if !success {
		fmt.Printf("Failed to start CPU Profiler\n")
		return
	}

	success, err = admin.StopCPUProfiler()
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		return
	}

	if !success {
		fmt.Printf("Failed to stop CPU Profiler\n")
		return
	}
}
