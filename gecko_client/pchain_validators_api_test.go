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
	validators, err := client.PChainApi().GetCurrentValidators()
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, len(validators), 5)
}
