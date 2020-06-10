package gecko_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

// ============= RPC Requester ===================
const (
	JSON_RPC_VERSION = "2.0"
)

type geckoJsonRpcRequester struct {
	ipAddr string
	port nat.Port
}

// This needs to be public so the JSON package can serialize it
type JsonRpcRequest struct {
	JsonRpc string	`json:"jsonrpc"`
	Method string 	`json:"method"`
	Params map[string]interface{} `json:"params"`
	Id int `json:"id"`
}

func (requester geckoJsonRpcRequester) makeRpcRequest(endpoint string, method string, params map[string]interface{}) ([]byte, error) {
	// Either Golang or Ava have a very nasty & subtle behaviour where duplicated '//' in the URL is treated as GET, even if it's POST
	// https://stackoverflow.com/questions/23463601/why-golang-treats-my-post-request-as-a-get-one
	endpoint = strings.TrimLeft(endpoint, "/")

	request := JsonRpcRequest{
		JsonRpc: JSON_RPC_VERSION,
		Method: method,
		Params:  params,
		// TODO let the user set this?
		Id: 1,
	}

	requestBodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Could not marshall request to endpoint '%v' with method '%v' and params '%v' to JSON",
			endpoint,
			method,
			params)
	}

	url := fmt.Sprintf("http://%v:%v/%v", requester.ipAddr, requester.port.Int(), endpoint)

	logrus.Tracef("Making request to url: %v", url)
	logrus.Tracef("Request body: %v", string(requestBodyBytes))
	resp, err := http.Post(
		url,
		"application/json",
		bytes.NewBuffer(requestBodyBytes),
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred when making JSON RPC POST request to %v", url)
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	logrus.Tracef("Got response with status code: %v", statusCode)

	responseBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred when reading response body")
	}
	logrus.Tracef("Response body: %v", string(responseBodyBytes))

	if statusCode != 200 {
		return nil, stacktrace.NewError(
			"Received response with non-200 code '%v' and response body '%v'",
			statusCode,
			string(responseBodyBytes))
	}
	return responseBodyBytes, nil
}
