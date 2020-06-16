package gecko_client

import (
	"encoding/json"
	"github.com/palantir/stacktrace"
)

// A struct that implements the jsonRpcRequester interface but just returns the same thing every time
type mockedJsonRpcRequester struct {
	resultStr string
}

func newMockedJsonRpcRequester(resultStr string) *mockedJsonRpcRequester {
	return &mockedJsonRpcRequester{resultStr: resultStr}
}

func (requester mockedJsonRpcRequester) makeRpcRequest(endpoint string, method string, params map[string]interface{}) ([]byte, error) {
	bytes := []byte(requester.resultStr)
	return bytes, nil
}

func (requester mockedJsonRpcRequester) makeUnmarshalledRpcRequest(endpoint string, method string, params map[string]interface{}, response *interface{}) error {
	bytes := []byte(requester.resultStr)
	if err := json.Unmarshal(bytes, response); err != nil {
		return stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return nil
}
