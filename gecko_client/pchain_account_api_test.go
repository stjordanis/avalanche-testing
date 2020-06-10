package gecko_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "address": "Q4MzFZZDPHRPAHFeDs3NiyyaZDvxHKivf"
    },
    "id": 1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	address, err := client.PChainApi().CreateAccount(
		"bob",
		"loblaw",
		"24jUJ9vZexUM6expyMcT48LBx27k1m7xpraoV62oSQAHdziao5")
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, "Q4MzFZZDPHRPAHFeDs3NiyyaZDvxHKivf", address)
}

func TestImportKey(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :3,
    "result" :{
        "address":"7u5FQArVaMSgGZzeTE9ckheWtDhU5T3KS"
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	address, err := client.PChainApi().ImportKey(
		"bob",
		"loblaw",
		"2w4XiXxPfQK4TypYqnohRL8DRNTz9cGiGmwQ1zmgEqD9c9KWLq")
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, "7u5FQArVaMSgGZzeTE9ckheWtDhU5T3KS", address)
}

func TestExportKey(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :3,
    "result" :{
        "privateKey":"2w4XiXxPfQK4TypYqnohRL8DRNTz9cGiGmwQ1zmgEqD9c9KWLq"
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	address, err := client.PChainApi().ExportKey(
		"bob",
		"loblaw",
		"7u5FQArVaMSgGZzeTE9ckheWtDhU5T3KS")
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, "2w4XiXxPfQK4TypYqnohRL8DRNTz9cGiGmwQ1zmgEqD9c9KWLq", address)
}

func TestGetAccount(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "address": "NcbCRXGMpHxukVmT8sirZcDnCLh1ykWp4",
        "nonce": "0",
        "balance": "0"
    },
    "id": 84
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	accountInfo, err := client.PChainApi().GetAccount("NcbCRXGMpHxukVmT8sirZcDnCLh1ykWp4")
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, "NcbCRXGMpHxukVmT8sirZcDnCLh1ykWp4", accountInfo.Address)
	assert.Equal(t, "0", accountInfo.Nonce)
	assert.Equal(t, "0", accountInfo.Balance)
}

func TestListAccounts(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "accounts": [
            {
                "address": "Q4MzFZZDPHRPAHFeDs3NiyyaZDvxHKivf",
                "nonce": "0",
                "balance": "0"
            },
            {
                "address": "NcbCRXGMpHxukVmT8sirZcDnCLh1ykWp4",
                "nonce": "0",
                "balance": "0"
            }
        ]
    },
    "id": 1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	accounts, err := client.PChainApi().ListAccounts("bob", "loblaw")
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, 2, len(accounts))
}

