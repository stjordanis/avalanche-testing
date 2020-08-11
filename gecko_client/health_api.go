package gecko_client

import (
	"encoding/json"

	"github.com/palantir/stacktrace"

	"github.com/ava-labs/gecko/api/health"
)

const (
	healthApiEndpoint = "ext/health"
)

type HealthApi struct {
	rpcRequester jsonRpcRequester
}

func (api HealthApi) GetLiveness() (health.GetLivenessReply, error) {
	var response health.GetLivenessReply
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(healthApiEndpoint, "health.getLiveness", make(map[string]interface{}))
	if err != nil {
		return health.GetLivenessReply{}, stacktrace.Propagate(err, "Error getting liveness")
	}

	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return health.GetLivenessReply{}, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response, nil
}
