package gecko_client

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
