package main

import (
	// "github.com/kurtosis-tech/kurtosis/ava_commons/testsuite"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	testName := os.Args[1]
	networkInfoFilepath := os.Args[2]
    println(fmt.Sprintf("Would run %v:" , testName))

	data, err := ioutil.ReadFile(networkInfoFilepath)
	if err != nil {
		// TODO make this a proper error
		panic("Could not read file bytes!")
	}
	println(fmt.Sprintf("Contents of file: %v", string(data)))
}
