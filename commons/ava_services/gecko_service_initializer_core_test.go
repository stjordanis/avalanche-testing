package ava_services

import (
	"bytes"
	"fmt"
	"net"
	"testing"

	"github.com/ava-labs/avalanche-e2e-tests/commons/ava_services/cert_providers"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/stretchr/testify/assert"
)

var testPublicIp = net.ParseIP("172.17.0.2")

func TestNoDepsStartCommand(t *testing.T) {
	initializerCore := NewGeckoServiceInitializerCore(
		1,
		1,
		false,
		make(map[string]string),
		[]string{},
		cert_providers.NewStaticGeckoCertProvider(bytes.Buffer{}, bytes.Buffer{}),
		LOG_LEVEL_INFO)

	expected := []string{
		avalancheBinary,
		"--public-ip=" + testPublicIp.String(),
		"--network-id=local",
		"--http-port=9650",
		"--http-host=",
		"--staking-port=9651",
		"--log-level=info",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-enabled=false",
		"--avax-tx-fee=0", // TODO make this into a variable
	}
	actual, err := initializerCore.GetStartCommand(make(map[string]string), testPublicIp, make([]services.Service, 0))
	assert.NoError(t, err, "An error occurred getting the start command")
	assert.Equal(t, expected, actual)
}

func TestWithDepsStartCommand(t *testing.T) {
	testNodeId := "node1"
	testDependencyIp := "1.2.3.4"

	bootstrapperNodeIDs := []string{
		testNodeId,
	}
	initializerCore := NewGeckoServiceInitializerCore(
		1,
		1,
		false,
		make(map[string]string),
		bootstrapperNodeIDs,
		cert_providers.NewStaticGeckoCertProvider(bytes.Buffer{}, bytes.Buffer{}),
		LOG_LEVEL_INFO)

	expected := []string{
		avalancheBinary,
		"--public-ip=" + testPublicIp.String(),
		"--network-id=local",
		"--http-port=9650",
		"--http-host=",
		"--staking-port=9651",
		"--log-level=info",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-enabled=false",
		"--avax-tx-fee=0",
		fmt.Sprintf("--bootstrap-ips=%v:9651", testDependencyIp),
	}

	testDependency := GeckoService{
		ipAddr:      "1.2.3.4",
		jsonRpcPort: "9650/tcp",
		stakingPort: "9651/tcp",
	}
	testDependencySlice := []services.Service{
		testDependency,
	}
	actual, err := initializerCore.GetStartCommand(make(map[string]string), testPublicIp, testDependencySlice)
	assert.NoError(t, err, "An error occurred getting the start command")
	assert.Equal(t, expected, actual)
}
