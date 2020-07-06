package gecko_client

// A struct that implements the JsonRpcRequester interface but just returns the same thing every time
type mockedJsonRpcRequester struct {
	resultStr string
}

func (requester mockedJsonRpcRequester) MakeRpcRequest(endpoint string, method string, params map[string]interface{}) ([]byte, error) {
	bytes := []byte(requester.resultStr)
	return bytes, nil
}
