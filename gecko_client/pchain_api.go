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

// Gets information about the account specified by the given address
func (api PChainApi) GetAccount(address string) (AccountInfo, error) {
	params := map[string]interface{}{
		"address": address,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.getAccount", params)
	if err != nil {
		return AccountInfo{}, stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response GetAccountResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return AccountInfo{}, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result, nil
}

// List accounts controlled by the user identified by the given username and password
func (api PChainApi) ListAccounts(username string, password string) ([]AccountInfo, error) {
	params := map[string]interface{}{
		"username": username,
		"password": password,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.listAccounts", params)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response ListAccountsResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return nil, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Accounts, nil
}




// ============= Validators ====================

// Gets the list of current validators that the Gecko node knows about
// A nil subnetId pointer will not send the parameter (which at time of writing means "use the default subnet")
func (api PChainApi) GetCurrentValidators(subnetIdPtr *string) ([]Validator, error) {
	params := map[string]interface{}{}
	if subnetIdPtr != nil {
		params["subnetID"] = *subnetIdPtr
	}

	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.getCurrentValidators", params)
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

// A nil subnetId pointer will not send the parameter (which at time of writing means "use the default subnet")
func (api PChainApi) GetPendingValidators(subnetIdPtr *string) ([]Validator, error) {
	params := map[string]interface{}{}
	if subnetIdPtr != nil {
		params["subnetID"] = *subnetIdPtr
	}

	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.getPendingValidators", params)
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

// A nil subnetId pointer will not send the parameter (which at time of writing means "use the default subnet")
func (api PChainApi) SampleValidators(subnetIdPtr *string) ([]string, error) {
	params := map[string]interface{}{}
	if subnetIdPtr != nil {
		params["subnetID"] = *subnetIdPtr
	}

	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.sampleValidators", params)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response SampleValidatorsResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return nil, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Validators, nil
}

func (api PChainApi) AddDefaultSubnetValidator(
		id string,
		startTime int64,
		endTime int64,
		stakeAmount int64,
		payerNonce int,
		destination string,
		delegationFeeRate int64) (string, error) {
	params := map[string]interface{}{
		"id": id,
		"payerNonce": payerNonce,
		"destination": destination,
		"startTime": startTime,
		"endTime": endTime,
		"stakeAmount": stakeAmount,
		"delegationFeeRate": delegationFeeRate,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.addDefaultSubnetValidator", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response AddDefaultSubnetValidatorResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.UnsignedTx, nil
}

func (api PChainApi) AddNonDefaultSubnetValidator(
		id string,
		subnetID string,
		startTime int64,
		endTime int64,
		weight int,
		payerNonce int) (string, error) {
	params := map[string]interface{}{
		"id": id,
		"subnetID": subnetID,
		"startTime": startTime,
		"endTime": endTime,
		"weight": weight,
		"payerNonce": payerNonce,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.addNonDefaultSubnetValidator", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response AddNonDefaultSubnetValidatorResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.UnsignedTx, nil
}

func (api PChainApi) AddDefaultSubnetDelegator(
		id string,
		startTime int64,
		endTime int64,
		stakeAmount int64,
		payerNonce int,
		destination string) (string, error) {
	params := map[string]interface{}{
		"id": id,
		"payerNonce": payerNonce,
		"destination": destination,
		"startTime": startTime,
		"endTime": endTime,
		"stakeAmount": stakeAmount,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.addDefaultSubnetDelegator", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response AddDefaultSubnetDelegator
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.UnsignedTx, nil
}

// =============== Subnets =====================


// Create an unsigned transaction to create a new Subnet.
func (api PChainApi) CreateSubnet(controlKeys []string, threshold int, payerNonce int) (string, error) {
	params := map[string]interface{}{
		"controlKeys": controlKeys,
		"threshold": threshold,
		"payerNonce": payerNonce,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.createSubnet", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response CreateUnsignedTransactionResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.UnsignedTx, nil
}


func (api PChainApi) GetSubnets() ([]Subnet, error) {
	params := map[string]interface{}{}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.getSubnets", params)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response GetSubnetsResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return nil, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Subnets, nil
}


func (api PChainApi) ValidatedBy(blockchainId string) (string, error) {
	params := map[string]interface{}{
		"blockchainID": blockchainId,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.validatedBy", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response ValidatedByResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.SubnetID, nil
}



func (api PChainApi) Validates(subnetId string) ([]string, error) {
	params := map[string]interface{}{
		"subnetID": subnetId,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.validates", params)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response ValidatesResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return nil, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.BlockchainIDs, nil
}


func (api PChainApi) GetBlockchains() ([]Blockchain, error) {
	params := map[string]interface{}{}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.getBlockchains", params)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response GetBlockchainsResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return nil, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Blockchains, nil
}


func (api PChainApi) ExportAVA(amount int, to string, payerNonce int) (string, error) {
	params := map[string]interface{}{
		"amount": amount,
		"to": to,
		"payerNonce": payerNonce,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.exportAVA", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response ExportAVAResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.UnsignedTx, nil
}


func (api PChainApi) ImportAVA(username string, password string, to string, payerNonce int) (string, error) {
	params := map[string]interface{}{
		"username": username,
		"password": password,
		"to": to,
		"payerNonce": payerNonce,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.importAVA", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response ImportAVAResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Tx, nil
}

func (api PChainApi) Sign(tx string, signer string, username string, password string) (string, error) {
	params := map[string]interface{}{
		"tx": tx,
		"signer": signer,
		"username": username,
		"password": password,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.sign", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response SignResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Tx, nil
}


func (api PChainApi) IssueTx(tx string) (string, error) {
	params := map[string]interface{}{
		"tx": tx,
	}
	var response interface{}
	err := api.rpcRequester.makeUnmarshalledRpcRequest(pchainEndpoint, "platform.issueTx", params, &response)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}
	castedResponse := response.(IssueTxResponse)
	return castedResponse.Result.TxID, nil
}
