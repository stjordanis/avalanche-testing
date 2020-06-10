package gecko_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBlockchainStatus(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "status": "Created"
    },
    "id": 1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	status, err := client.PChainApi().GetBlockchainStatus("test-blockchain-id")
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, status, "Created")
}