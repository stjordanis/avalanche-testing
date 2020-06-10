package gecko_client

import (
	"encoding/json"
	"github.com/palantir/stacktrace"
)

const (
	pchainEndpoint = "ext/P"
)

type PChainApi struct {
	rpcRequester jsonRpcRequester
}

// ============= Blockchain ====================
func (api PChainApi) CreateBlockchain(vmId string, subnetId string, name string, genesisData string, payerNonce int) (string, error) {
	params := map[string]interface{}{
		"vmID": vmId,
		"SubnetID": subnetId,
		"name": name,
		"genesisData": genesisData,
		"payerNonce" : payerNonce,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.createBlockchain", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response CreateBlockchainResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.UnsignedTx, nil
}

func (api PChainApi) GetBlockchainStatus(blockchainId string) (string, error) {
	params := map[string]interface{}{
		"blockchainID": blockchainId,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.getBlockchainStatus", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response GetBlockchainStatusResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Status, nil
}

// ============= Validators ====================
func (api PChainApi) GetCurrentValidators() ([]Validator, error) {
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.getCurrentValidators", make(map[string]interface{}))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response GetValidatorsResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return nil, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Validators, nil
}
