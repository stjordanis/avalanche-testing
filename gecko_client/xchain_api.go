package gecko_client

import (
	"encoding/json"
	"github.com/palantir/stacktrace"
)

const (
	xchainEndpoint = "ext/bc/X"
)

type XChainApi struct {
	rpcRequester jsonRpcRequester
}

func (api XChainApi) ImportKey(username string, password string, privateKey string) (string, error) {
	params := map[string]interface{}{
		"username": username,
		"password": password,
		"privateKey": privateKey,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(xchainEndpoint, "avm.importKey", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	var response ImportKeyResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Address, nil
}

func (api XChainApi) ExportAVA(to string, amount int64, username string, password string) (string, error) {
	params := map[string]interface{}{
		"to": to,
		"amount": amount,
		"username": username,
		"password": password,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(xchainEndpoint, "avm.exportAVA", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	var response XChainExportAVAResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.TxID, nil
}

func (api XChainApi) GetTxStatus(txnId string) (string, error) {
	params := map[string]interface{}{
		"txID": txnId,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(xchainEndpoint, "avm.getTxStatus", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	var response GetTxStatusResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Status, nil
}

func (api XChainApi) GetBalance(address string, assetId string) (*AccountWithUtxoInfo, error) {
	params := map[string]interface{}{
		"address": address,
		"assetID": assetId,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(xchainEndpoint, "avm.getBalance", params)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error making request")
	}

	var response GetBalanceResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return nil, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return &response.Result, nil
}

func (api XChainApi) Send(amount int64, assetId string, to string, username string, password string) (string, error) {
	params := map[string]interface{}{
		"amount": amount,
		"assetID": assetId,
		"to": to,
		"username": username,
		"password": password,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(xchainEndpoint, "avm.send", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	var response SendResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.TxID, nil
}

func (api XChainApi) CreateAddress(username string, password string) (string, error) {
	params := map[string]interface{}{
		"username": username,
		"password": password,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(xchainEndpoint, "avm.createAddress", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	var response CreateAddressResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Address, nil
}

func (api XChainApi) IssueTx(tx string) (string, error) {
	params := map[string]interface{}{
		"tx": tx,
	}
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(xchainEndpoint, "avm.issueTx", params)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error making request")
	}

	var response IssueTxResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return "", stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.TxID, nil
}
