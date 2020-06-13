package gecko_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCurrentValidators(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "validators": [
            {
                "startTime": "1572566400",
                "endtime": "1604102400",
                "stakeAmount": "20000000000000",
                "id": "MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ"
            },
            {
                "startTime": "1572566400",
                "endtime": "1604102400",
                "stakeAmount": "20000000000000",
                "id": "GWPcbFJZFfZreETSoWjPimr846mXEKCtu"
            },
            {
                "startTime": "1572566400",
                "endtime": "1604102400",
                "stakeAmount": "20000000000000",
                "id": "NFBbbJ4qCmNaCzeW7sxErhvWqvEQMnYcN"
            },
            {
                "startTime": "1572566400",
                "endtime": "1604102400",
                "stakeAmount": "20000000000000",
                "id": "7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg"
            },
            {
                "startTime": "1572566400",
                "endtime": "1604102400",
                "stakeAmount": "20000000000000",
                "id": "P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5"
            }
        ]
    },
    "id": 85
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	validators, err := client.PChainApi().GetCurrentValidators(nil)
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, len(validators), 5)
}

func TestGetPendingValidators(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "validators": [
            {
                "startTime": "1572567400",
                "endtime": "1604102400",
                "stakeAmount": "20000000000000",
                "id": "MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ"
            }
        ]
    },
    "id": 1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	validators, err := client.PChainApi().GetPendingValidators(nil)
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, 1, len(validators))
}

func TestSampleValidators(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :1,
    "result" :{
        "validators":[
            "MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ",
            "NFBbbJ4qCmNaCzeW7sxErhvWqvEQMnYcN"
        ]
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	validatorIds, err := client.PChainApi().SampleValidators(nil)
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, 2, len(validatorIds))
}

func TestAddDefaultSubnetValidator(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :1,
    "result" :{
        "unsignedTx": "1115K3jV5Yxr145wi6kEYpN1nPz3GEBkzG8mpF2s2959VsR54YGenLJrgdg3UEE7vFPNDE5n3Cq9Vs71HEjUUoVSyrt9Z3X7M5sKLCX5WScTcQocxjnXfFowZxFe4uH8iJU7jnCZgeKK5bWsfnWy2b9PbCQMN2uNLvwyKRp4ZxcgRptkuXRMCKHfhbHVKBYmr5e2VbBBht19be57uFUP5yVdMxKnxecs"
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	unsignedTx, err := client.PChainApi().AddDefaultSubnetValidator(
		"ARCLrphAHZ28xZEBfUL7SVAmzkTZNe1LK",
		1591837350,
		1591920000,
		1000000,
		1,
		"Q4MzFZZDPHRPAHFeDs3NiyyaZDvxHKivf",
		100000)
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(
		t,
		"1115K3jV5Yxr145wi6kEYpN1nPz3GEBkzG8mpF2s2959VsR54YGenLJrgdg3UEE7vFPNDE5n3Cq9Vs71HEjUUoVSyrt9Z3X7M5sKLCX5WScTcQocxjnXfFowZxFe4uH8iJU7jnCZgeKK5bWsfnWy2b9PbCQMN2uNLvwyKRp4ZxcgRptkuXRMCKHfhbHVKBYmr5e2VbBBht19be57uFUP5yVdMxKnxecs",
		unsignedTx)
}

func TestAddNonDefaultSubnetValidator(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :1,
    "result" :{
        "unsignedTx": "1115K3jV5Yxr145wi6kEYpN1nPz3GEBkzG8mpF2s2959VsR54YGenLJrgdg3UEE7vFPNDE5n3Cq9Vs71HEjUUoVSyrt9Z3X7M5sKLCX5WScTcQocxjnXfFowZxFe4uH8iJU7jnCZgeKK5bWsfnWy2b9PbCQMN2uNLvwyKRp4ZxcgRptkuXRMCKHfhbHVKBYmr5e2VbBBht19be57uFUP5yVdMxKnxecs"
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	unsignedTx, err := client.PChainApi().AddNonDefaultSubnetValidator(
		"7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg",
		"zBfoWW1FfkPVRfywpJ1CVQRfnYesEpdFC61hmU2n9JNGhDUEL",
		1583524047,
		1604102399,
		1,
		2)
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(
		t,
		"1115K3jV5Yxr145wi6kEYpN1nPz3GEBkzG8mpF2s2959VsR54YGenLJrgdg3UEE7vFPNDE5n3Cq9Vs71HEjUUoVSyrt9Z3X7M5sKLCX5WScTcQocxjnXfFowZxFe4uH8iJU7jnCZgeKK5bWsfnWy2b9PbCQMN2uNLvwyKRp4ZxcgRptkuXRMCKHfhbHVKBYmr5e2VbBBht19be57uFUP5yVdMxKnxecs",
		unsignedTx)
}

func TestAddDefaultSubnetDelegator(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :1,
    "result" :{
        "unsignedTx": "111Bit5JNASbJyTLrd2kWkYRoc96swEWoWdmEhuGAFK3rCAyTnTzomuFwgx1SCUdUE71KbtXPnqj93KGr3CeftpPN37kVyqBaAQ5xaDjr7wU8riGS89NDJ8AwVgZgnFkgF3uMfwCiCuPvvubGyQxNHE4TM9iDgj6h3URdGQ4JntP44wokCEP3ADn7sMM8kUTbmcNo84U87"
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	unsignedTx, err := client.PChainApi().AddDefaultSubnetDelegator(
		"MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ",
		1594102400,
		1604102400,
		100000,
		1,
		"Q4MzFZZDPHRPAHFeDs3NiyyaZDvxHKivf")
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(
		t,
		"111Bit5JNASbJyTLrd2kWkYRoc96swEWoWdmEhuGAFK3rCAyTnTzomuFwgx1SCUdUE71KbtXPnqj93KGr3CeftpPN37kVyqBaAQ5xaDjr7wU8riGS89NDJ8AwVgZgnFkgF3uMfwCiCuPvvubGyQxNHE4TM9iDgj6h3URdGQ4JntP44wokCEP3ADn7sMM8kUTbmcNo84U87",
		unsignedTx)
}

