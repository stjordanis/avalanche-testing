package gecko_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateBlockchain(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "unsignedTx": "111498J8u7uGkNzTKn2r7QUDPC1gq3Hb9XvAVNYHBK8AG2NXVMqo54SyiVAGFm1Ax5vGZgmxbuAMRS1TfsemkVDwK5N2Y5NzgU3pkT2WG9vJgg1N4m6gmDQp3WrKTa94eFWF4kwnjgAa8dLPBvFViCRY5FBtVAj3bXxMVPxYCn1THakh4dVmnHycQsdB3Hds3GHxQmYSXW712qHEvt2p4pd2Rk2grqAgvXLSgha1X3iovaeRM93KQiasYx8VTynPNwMmEo4NPs4x6GgEiSbGdxg9wRTcByG"
    },
    "id": 1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	unsignedTx, err := client.PChainApi().CreateBlockchain(
		"timestamp",
		"2bRCr6B4MiEfSjidDwxDpdCyviwnfUVqB2HGwhm947w9YYqb7r",
		"My new timestamp",
		"45oj4CqFViNHUtBxJ55TZfqaVAXFwMRMj2XkHVqUYjJYoTaEM",
		6)
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t,
		"111498J8u7uGkNzTKn2r7QUDPC1gq3Hb9XvAVNYHBK8AG2NXVMqo54SyiVAGFm1Ax5vGZgmxbuAMRS1TfsemkVDwK5N2Y5NzgU3pkT2WG9vJgg1N4m6gmDQp3WrKTa94eFWF4kwnjgAa8dLPBvFViCRY5FBtVAj3bXxMVPxYCn1THakh4dVmnHycQsdB3Hds3GHxQmYSXW712qHEvt2p4pd2Rk2grqAgvXLSgha1X3iovaeRM93KQiasYx8VTynPNwMmEo4NPs4x6GgEiSbGdxg9wRTcByG",
		unsignedTx)
}

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
	assert.Equal(t, "Created", status)
}