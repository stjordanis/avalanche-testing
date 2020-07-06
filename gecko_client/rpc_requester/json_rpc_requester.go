package rpc_requester

type JsonRpcRequester interface {
	MakeRpcRequest(endpoint string, method string, params map[string]interface{}) ([]byte, error)
}
