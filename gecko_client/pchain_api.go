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

// Creates a blockchain with the given parameters, returning the unsigned transaction identifier
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

// Gets the status of the given blockchain ID
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

// ============= Accounts ====================

// Creates an account with the given parameters, returning the account address
func (api PChainApi) CreateAccount(username string, password string, privateKey string) (string, error) {
	params := map[string]interface{}{
		"username": username,
		"password": password,
		"privateKey": privateKey,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.createAccount", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response CreateAccountResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Address, nil
}

func (api PChainApi) ImportKey(username string, password string, privateKey string) (string, error) {
	params := map[string]interface{}{
		"username": username,
		"password": password,
		"privateKey": privateKey,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.importKey", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response ImportKeyResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Address, nil
}

// Returns the private key associated with the given username, password, and address
func (api PChainApi) ExportKey(username string, password string, address string) (string, error) {
	params := map[string]interface{}{
		"username" : username,
		"password": password,
		"address": address,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.exportKey", params)

	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response ExportKeyResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.PrivateKey, nil
}



// ============= Validators ====================

// Gets the list of current validators that the Gecko node knows about
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
