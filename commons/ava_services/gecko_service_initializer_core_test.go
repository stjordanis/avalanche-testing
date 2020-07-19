package ava_services

import (
	"bytes"
	"fmt"
	"github.com/kurtosis-tech/ava-e2e-tests/commons/ava_services/cert_providers"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/stretchr/testify/assert"
	"testing"
)


const TEST_PUBLIC_IP ="172.17.0.2"


func TestNoDepsStartCommand(t *testing.T) {
	initializerCore := NewGeckoServiceInitializerCore(
		1,
		1,
		false,
		[]string{},
		cert_providers.NewStaticGeckoCertProvider(bytes.Buffer{}, bytes.Buffer{}),
		LOG_LEVEL_INFO)

	expected := []string{
		"/gecko/build/ava",
		"--public-ip=" + TEST_PUBLIC_IP,
		"--network-id=local",
		"--http-port=9650",
		"--http-host=" + TEST_PUBLIC_IP,
		"--staking-port=9651",
		"--log-level=info",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-tls-enabled=false",
	}
	actual, err := initializerCore.GetStartCommand(make(map[string]string), TEST_PUBLIC_IP, make([]services.Service, 0))
	assert.NoError(t, err, "An error occurred getting the start command")
	assert.Equal(t, expected, actual)
}

func TestWithDepsStartCommand(t *testing.T) {
	testNodeId := "node1"
	testDependencyIp := "1.2.3.4"

	bootstrapperNodeIds := []string{
		testNodeId,
	}
	initializerCore := NewGeckoServiceInitializerCore(
		1,
		1,
		false,
		bootstrapperNodeIds,
		cert_providers.NewStaticGeckoCertProvider(bytes.Buffer{}, bytes.Buffer{}),
		LOG_LEVEL_INFO)

	expected := []string{
		"/gecko/build/ava",
		"--public-ip=" + TEST_PUBLIC_IP,
		"--network-id=local",
		"--http-port=9650",
		"--http-host=" + TEST_PUBLIC_IP,
		"--staking-port=9651",
		"--log-level=info",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-tls-enabled=false",
		fmt.Sprintf("--bootstrap-ips=%v:9651", testDependencyIp),
	}

	testDependency := GeckoService{
		ipAddr: "1.2.3.4",
		jsonRpcPort: "9650/tcp",
		stakingPort: "9651/tcp",
	}
	testDependencySlice := []services.Service{
		testDependency,
	}
	actual, err := initializerCore.GetStartCommand(make(map[string]string), TEST_PUBLIC_IP, testDependencySlice)
	assert.NoError(t, err, "An error occurred getting the start command")
	assert.Equal(t, expected, actual)
}
