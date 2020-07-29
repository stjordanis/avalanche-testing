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
		// {
		// 	resultString: `{
		//       "jsonrpc":"2.0",
		//          "result":{
		//             "checks":{
		//                "chains.default.bootstrapped":{
		//                   "message":"didn't run yet",
		//                   "error":{
		//                      "message":"didn't run yet"
		//                   },
		//                   "timestamp":"2020-07-28T14:33:48.140525-04:00",
		//                   "contiguousFailures":1,
		//                   "timeOfFirstFailure":"2020-07-28T14:33:48.140525-04:00"
		//                },
		//                "network.validators.heartbeat":{
		//                   "message":"didn't run yet",
		//                   "error":{
		//                      "message":"didn't run yet"
		//                   },
		//                   "timestamp":"2020-07-28T14:33:48.140501-04:00",
		//                   "contiguousFailures":1,
		//                   "timeOfFirstFailure":"2020-07-28T14:33:48.140501-04:00"
		//                }
		//          },
		//          "healthy":false
		//       },
		//       "id":1
		//    }`,
		// 	expectedChecks:  2,
		// 	expectedHealthy: false,
		// 	nilError:        true,
		// },
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
