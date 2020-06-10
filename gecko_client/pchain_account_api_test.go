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
	assert.Equal(t, address, "Q4MzFZZDPHRPAHFeDs3NiyyaZDvxHKivf")
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
	assert.Equal(t, address, "7u5FQArVaMSgGZzeTE9ckheWtDhU5T3KS")
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
	assert.Equal(t, address, "2w4XiXxPfQK4TypYqnohRL8DRNTz9cGiGmwQ1zmgEqD9c9KWLq")
}

