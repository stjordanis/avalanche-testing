package gecko_client

import (
	"encoding/json"
	"github.com/kurtosis-tech/ava-e2e-tests/gecko_client/rpc_requester"
	"github.com/palantir/stacktrace"
)

const (
	keystoreEndpoint = "ext/keystore"
)

type KeystoreApi struct {
	rpcRequester rpc_requester.JsonRpcRequester
}

// Creates a blockchain with the given parameters, returning the unsigned transaction identifier
func (api KeystoreApi) CreateUser(username string, password string) (bool, error) {
	params := map[string]interface{}{
		"username": username,
		"password": password,
	}
	responseBodyBytes, err := api.rpcRequester.MakeRpcRequest(keystoreEndpoint, "keystore.createUser", params)
	if err != nil {
		return false, stacktrace.Propagate(err, "Error making request")
	}

	var response CreateUserResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return false, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Success, nil
}
