package gecko_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLiveness(t *testing.T) {
	resultStr := `{
   "jsonrpc":"2.0",
   "result":{
      "checks":{
         "network.validators.heartbeat":{
            "message":{
               "heartbeat":1591041377
            },
            "timestamp":"2020-06-01T15:56:18.554202-04:00",
            "duration":23201,
            "contiguousFailures":0,
            "timeOfFirstFailure":null
         }
      },
      "healthy":true
   },
   "id":1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	livenessInfo, err := client.HealthApi().GetLiveness()
	assert.Nil(t, err, "Error message should be nil")
	assert.True(t, livenessInfo.Healthy)

	checks := livenessInfo.Checks
	assert.Equal(t, len(checks), 1)
}
