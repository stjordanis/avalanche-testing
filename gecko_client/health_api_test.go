package gecko_client

import (
	"testing"
)

func TestGetLiveness(t *testing.T) {
	tests := []struct {
		resultString    string
		expectedChecks  int
		expectedHealthy bool
		nilError        bool
	}{
		{
			resultString: `{
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
         }`,
			expectedChecks:  1,
			expectedHealthy: true,
			nilError:        true,
		},
		// TODO fix health API to not error when parsing health check that contains an error
	}

	for _, test := range tests {
		client := clientFromRequester(mockedJsonRpcRequester{resultStr: test.resultString})
		livenessInfo, err := client.HealthApi().GetLiveness()

		if test.nilError && err != nil {
			t.Fatalf("Expected error to be nil, but found: %w", err)
		}

		if test.expectedChecks != len(livenessInfo.Checks) {
			t.Fatalf("Expected to find: %d checks, but instead found %d checks", test.expectedChecks, len(livenessInfo.Checks))
		}

		if test.expectedHealthy && !livenessInfo.Healthy {
			t.Fatal("Unexpectedly returned unhealthy")
		}
	}
}
