package gecko_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestXChainImportKey(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :1,
    "result" :{
        "address":"X-7u5FQArVaMSgGZzeTE9ckheWtDhU5T3KS"
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	address, err := client.XChainApi().ImportKey("myUsername", "myPassword", "2w4XiXxPfQK4TypYqnohRL8DRNTz9cGiGmwQ1zmgEqD9c9KWLq")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, "X-7u5FQArVaMSgGZzeTE9ckheWtDhU5T3KS", address)
}

func TestXChainExportAva(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "txID": "25VzbNzt3gi2vkE3Kr6H9KJeSR2tXkr8FsBCm3vARnB5foLVmx"
    },
    "id": 1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	transactionId, err := client.XChainApi().ExportAVA("Bg6e45gxCUTLXcfUuoy3go2U6V3bRZ5jH", 500, "myUsername", "myPassword")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, "25VzbNzt3gi2vkE3Kr6H9KJeSR2tXkr8FsBCm3vARnB5foLVmx", transactionId)
}

func TestXChainImportAva(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "txID": "MqEaeWc4rfkw9fhRMuMTN7KUTNpFmh9Fd7KSre1ZqTsTQG73h"
    },
    "id": 1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	transactionId, err := client.XChainApi().ImportAVA("X-G5ZGXEfoWYNFZH5JF9C4QPKAbPTKwRbyB", "myUsername", "myPassword")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, "MqEaeWc4rfkw9fhRMuMTN7KUTNpFmh9Fd7KSre1ZqTsTQG73h", transactionId)
}

func TestGetTxStatus(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :1,
    "result" :{
        "status":"Accepted"
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	status, err := client.XChainApi().GetTxStatus("2QouvFWUbjuySRxeX5xMbNCuAaKWfbk5FeEa2JmoF85RKLk2dD")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, "Accepted", status)
}

func TestGetBalance(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :1,
    "result" :{
        "balance":"299999999999900",
        "utxoIDs":[
            {
                "txID":"WPQdyLNqHfiEKp4zcCpayRHYDVYuh1hqs9c1RqgZXS4VPgdvo",
                "outputIndex":1
            }
        ]
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	accountWithUtxoInfo, err := client.XChainApi().GetBalance("X-EKpEPX56YA1dsaHBsW8X5nGqNSwJ7JrWH", "2pYGetDWyKdHxpFxh2LHeoLNCH6H5vxxCxHQtFnnFaYxLsqtHC")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, "299999999999900", accountWithUtxoInfo.Balance)
	assert.Equal(t, 1, len(accountWithUtxoInfo.UtxoIDs))

	utxoId := accountWithUtxoInfo.UtxoIDs[0]
	assert.Equal(t, 1, utxoId.OutputIndex)
	assert.Equal(t, "WPQdyLNqHfiEKp4zcCpayRHYDVYuh1hqs9c1RqgZXS4VPgdvo", utxoId.TxID)
}

func TestSend(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :1,
    "result" :{
        "txID":"2iXSVLPNVdnFqn65rRvLrsu8WneTFqBJRMqkBJx5vZTwAQb8c1"
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	transactionId, err := client.XChainApi().Send(10000, "AVA", "X-xMrKg8uUECt5CS9RE9j5hizv2t2SWTbk", "userThatControlsAtLeast10000OfThisAsset", "myPassword")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, "2iXSVLPNVdnFqn65rRvLrsu8WneTFqBJRMqkBJx5vZTwAQb8c1", transactionId)
}

func TestCreateAddress(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "address": "X-EKpEPX56YA1dsaHBsW8X5nGqNSwJ7JrWH"
    },
    "id": 1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	address, err := client.XChainApi().CreateAddress("myUsername", "myPassword")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, "X-EKpEPX56YA1dsaHBsW8X5nGqNSwJ7JrWH", address)
}

func TestXChainIssueTx(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :1,
    "result" :{
        "txID":"NUPLwbt2hsYxpQg4H2o451hmTWQ4JZx2zMzM4SinwtHgAdX1JLPHXvWSXEnpecStLj"
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	transactionId, err := client.XChainApi().IssueTx("6sTENqXfk3gahxkJbEPsmX9eJTEFZRSRw83cRJqoHWBiaeAhVbz9QV4i6SLd6Dek4eLsojeR8FbT3arFtsGz9ycpHFaWHLX69edJPEmj2tPApsEqsFd7wDVp7fFxkG6HmySR")
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, "NUPLwbt2hsYxpQg4H2o451hmTWQ4JZx2zMzM4SinwtHgAdX1JLPHXvWSXEnpecStLj", transactionId)
}


