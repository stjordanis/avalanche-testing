package gecko_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExportAVA(t *testing.T) {
	resultStr := `{
		"jsonrpc": "2.0",
		"result": {
			"unsignedTx": "1112Y8Y5ibRqMDtby9NSdpK9u3n1yGywybAAVYnhCkFYcRzEYbR7J5Ci6SX98PmgS2LpRf5pcu6YAgLYGiTuQpiSucRcX4dv7HbVnEsrQnjcieGbgkf9PFS126hC8xce4pEZUzr9jReVdfXe3g9BSUsXLj2XcWrnD6iTgHpiC18jjyjg1wjm1Vs4TcXhG472MRvGspucJ8LuUE91WV7353Kxdc2e7Trw2Sd6iV"
		},
		"id": 1
	}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	unsignedTx, err := client.PChainApi().ExportAVA(1,"G5ZGXEfoWYNFZH5JF9C4QPKAbPTKwRbyB", 2)
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(
		t,
		"1112Y8Y5ibRqMDtby9NSdpK9u3n1yGywybAAVYnhCkFYcRzEYbR7J5Ci6SX98PmgS2LpRf5pcu6YAgLYGiTuQpiSucRcX4dv7HbVnEsrQnjcieGbgkf9PFS126hC8xce4pEZUzr9jReVdfXe3g9BSUsXLj2XcWrnD6iTgHpiC18jjyjg1wjm1Vs4TcXhG472MRvGspucJ8LuUE91WV7353Kxdc2e7Trw2Sd6iV",
		unsignedTx)
}

func TestImportAVA(t *testing.T) {
	resultStr := `{
		"jsonrpc": "2.0",
		"result": {
			"tx": "1117xBwcr5fo1Ch4umyzjYgnuoFhSwBHdMCam2wRe8SxcJJvQRKSmufXM8aSqKaDmX4TjvzPaUbSn33TAQsbZDhzcHEGviuthncY5VQfUJogyMoFGXUtu3M8NbwNhrYtmSRkFdmN4w933janKvJYKNnsDMvMkmasxrFj8fQxE6Ej8eyU2Jqj2gnTxU2WD3NusFNKmPfgJs8DRCWgYyJVodnGvT43hovggVaWHHD8yYi9WJ64pLCvtCcEYkQeEeA5NE8eTxPtWJrwSMTciHHVdHMpxdVAY6Ptr2rMcYSacr8TZzw59XJfbQT4R6DCsHYQAPJAUfDNeX2JuiBk9xonfKmGcJcGXwdJZ3QrvHHHfHCeuxqS13AfU"
		},
		"id": 1
	}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	tx, err := client.PChainApi().ImportAVA(
		"bob",
		"loblaw",
		"Bg6e45gxCUTLXcfUuoy3go2U6V3bRZ5jH",
		1)
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t,
		"1117xBwcr5fo1Ch4umyzjYgnuoFhSwBHdMCam2wRe8SxcJJvQRKSmufXM8aSqKaDmX4TjvzPaUbSn33TAQsbZDhzcHEGviuthncY5VQfUJogyMoFGXUtu3M8NbwNhrYtmSRkFdmN4w933janKvJYKNnsDMvMkmasxrFj8fQxE6Ej8eyU2Jqj2gnTxU2WD3NusFNKmPfgJs8DRCWgYyJVodnGvT43hovggVaWHHD8yYi9WJ64pLCvtCcEYkQeEeA5NE8eTxPtWJrwSMTciHHVdHMpxdVAY6Ptr2rMcYSacr8TZzw59XJfbQT4R6DCsHYQAPJAUfDNeX2JuiBk9xonfKmGcJcGXwdJZ3QrvHHHfHCeuxqS13AfU",
		tx)
}


func TestSign(t *testing.T) {
	resultStr := `{
		"jsonrpc": "2.0",
		"result": {
			"Tx": "111Bit5JNASbJyTLrd2kWkYRoc96swEWoWdmEhuGAFK3rCAyTnTzomuFwgx1SCUdUE71KbtXPnqj93KGr3CeftpPN37kVyqBaAQ5xaDjr7wVBTUYi9iV7kYJnHF61yovViJF74mJJy7WWQKeRMDRTiPuii5gsd11gtNahCCsKbm9seJtk2h1wAPZn9M1eL84CGVPnLUiLP"
		},
		"id": 1
	}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	signTx, err := client.PChainApi().Sign(
		"111Bit5JNASbJyTLrd2kWkYRoc96swEWoWdmEhuGAFK3rCAyTnTzomuFwgx1SCUdUE71KbtXPnqj93KGr3CeftpPN37kVyqBaAQ5xaDjr7wU8riGS89NDJ8AwVgZgnFkgF3uMfwCiCuPvvubGyQxNHE4TM9iDgj6h3URdGQ4JntP44wokCEP3ADn7sMM8kUTbmcNo84U87",
		"6Y3kysjF9jnHnYkdS9yGAuoHyae2eNmeV",
		"bob",
		"loblaw")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(
		t,
		"111Bit5JNASbJyTLrd2kWkYRoc96swEWoWdmEhuGAFK3rCAyTnTzomuFwgx1SCUdUE71KbtXPnqj93KGr3CeftpPN37kVyqBaAQ5xaDjr7wVBTUYi9iV7kYJnHF61yovViJF74mJJy7WWQKeRMDRTiPuii5gsd11gtNahCCsKbm9seJtk2h1wAPZn9M1eL84CGVPnLUiLP",
		signTx)
}

func TestIssueTx(t *testing.T) {
	resultStr := `{
		"jsonrpc": "2.0",
		"result": {
			"txID": "G3BuH6ytQ2averrLxJJugjWZHTRubzCrUZEXoheG5JMqL5ccY"
		},
		"id": 1
	}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	txnId, err := client.PChainApi().IssueTx("111Bit5JNASbJyTLrd2kWkYRoc96swEWoWdmEhuGAFK3rCAyTnTzomuFwgx1SCUdUE71KbtXPnqj93KGr3CeftpPN37kVyqBaAQ5xaDjr7wVBTUYi9iV7kYJnHF61yovViJF74mJJy7WWQKeRMDRTiPuii5gsd11gtNahCCsKbm9seJtk2h1wAPZn9M1eL84CGVPnLUiLP")
	assert.Nil(t, err, "Error message should be nil")

	assert.Equal(t, txnId, "G3BuH6ytQ2averrLxJJugjWZHTRubzCrUZEXoheG5JMqL5ccY")
}
