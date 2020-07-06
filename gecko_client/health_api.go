package gecko_client

import (
	"encoding/json"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client/rpc_requester"
	"github.com/palantir/stacktrace"
)

const (
	healthApiEndpoint = "ext/health"
)


type HealthApi struct {
	rpcRequester rpc_requester.JsonRpcRequester
}

func (api HealthApi) GetLiveness() (LivenessInfo, error) {
	responseBodyBytes, err := api.rpcRequester.MakeRpcRequest(healthApiEndpoint, "health.getLiveness", make(map[string]interface{}))
	if err != nil {
		return LivenessInfo{}, stacktrace.Propagate(err, "Error getting liveness")
	}

	var response GetLivenessResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return LivenessInfo{}, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result, nil
}


