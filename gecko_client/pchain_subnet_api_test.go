package gecko_client

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTODO(t *testing.T) {
	testUnsignedTransactionResult := "1112LA7e8GvkGHDkxZa9Q7kszqvWHooumX5PhqA9NJG7erwXYcwQUPRQyukYX1ncu1DmWvvPNMuivUqvGp1t9M3wys5joqXrXtV2jescQ5AWaUKHiSBUWBRHseMLhGxWNT4Bv6LNVvaaA1ZW33avQBAzz7V84KpKGW7fD3Fz1okxknLgoG"
	resultStr := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"result": {
			"unsignedTx": "%s"
		},
		"id": 1
	}`, testUnsignedTransactionResult)
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	unsignedTxn, err := client.PChainApi().CreateSubnet([]string{"key1", "key2"}, 1, 1)
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, testUnsignedTransactionResult, unsignedTxn)
}


