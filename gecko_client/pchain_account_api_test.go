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
