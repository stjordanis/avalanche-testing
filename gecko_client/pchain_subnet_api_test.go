package gecko_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSubnet(t *testing.T) {
	resultStr := `{
		"jsonrpc": "2.0",
		"result": {
			"unsignedTx": "1112LA7e8GvkGHDkxZa9Q7kszqvWHooumX5PhqA9NJG7erwXYcwQUPRQyukYX1ncu1DmWvvPNMuivUqvGp1t9M3wys5joqXrXtV2jescQ5AWaUKHiSBUWBRHseMLhGxWNT4Bv6LNVvaaA1ZW33avQBAzz7V84KpKGW7fD3Fz1okxknLgoG"
		},
		"id": 1
	}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	unsignedTxn, err := client.PChainApi().CreateSubnet([]string{"key1", "key2"}, 1, 1)
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(
		t,
		"1112LA7e8GvkGHDkxZa9Q7kszqvWHooumX5PhqA9NJG7erwXYcwQUPRQyukYX1ncu1DmWvvPNMuivUqvGp1t9M3wys5joqXrXtV2jescQ5AWaUKHiSBUWBRHseMLhGxWNT4Bv6LNVvaaA1ZW33avQBAzz7V84KpKGW7fD3Fz1okxknLgoG",
		unsignedTxn)
}

func TestGetSubnets(t *testing.T) {
	resultStr := `{
		"jsonrpc": "2.0",
		"result": {
			"subnets": [
				{
					"id": "hW8Ma7dLMA7o4xmJf3AXBbo17bXzE7xnThUd3ypM4VAWo1sNJ",
					"controlKeys": [
						"KNjXsaA1sZsaKCD1cd85YXauDuxshTes2",
						"Aiz4eEt5xv9t4NCnAWaQJFNz5ABqLtJkR"
					],
					"threshold": "2"
				}
			]
		},
		"id": 6
	}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	subnetList, err := client.PChainApi().GetSubnets()
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, 1, len(subnetList))
	assert.Equal(t, 2, len(subnetList[0].ControlKeys))
	assert.Equal(t, "2", subnetList[0].Threshold)
	assert.Equal(
		t,
		"hW8Ma7dLMA7o4xmJf3AXBbo17bXzE7xnThUd3ypM4VAWo1sNJ",
		subnetList[0].Id)
}

func TestValidatedBy(t *testing.T) {
	resultStr := `{
		"jsonrpc": "2.0",
		"result": {
			"subnetID": "2bRCr6B4MiEfSjidDwxDpdCyviwnfUVqB2HGwhm947w9YYqb7r"
		},
		"id": 1
	}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	subnetId, err := client.PChainApi().ValidatedBy("KDYHHKjM4yTJTT8H8qPs5KXzE6gQH5TZrmP1qVr1P6qECj3XN")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t,"2bRCr6B4MiEfSjidDwxDpdCyviwnfUVqB2HGwhm947w9YYqb7r", subnetId)
}

func TestValidates(t *testing.T) {
	resultStr := `{
		"jsonrpc": "2.0",
		"result": {
			"blockchainIDs": [
				"KDYHHKjM4yTJTT8H8qPs5KXzE6gQH5TZrmP1qVr1P6qECj3XN",
				"2TtHFqEAAJ6b33dromYMqfgavGPF3iCpdG3hwNMiart2aB5QHi"
			]
		},
		"id": 1
	}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	blockchainIds, err := client.PChainApi().Validates("2bRCr6B4MiEfSjidDwxDpdCyviwnfUVqB2HGwhm947w9YYqb7r")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, 2, len(blockchainIds))
	assert.Equal(t, "KDYHHKjM4yTJTT8H8qPs5KXzE6gQH5TZrmP1qVr1P6qECj3XN", blockchainIds[0])
	assert.Equal(t, "2TtHFqEAAJ6b33dromYMqfgavGPF3iCpdG3hwNMiart2aB5QHi", blockchainIds[1])
}



