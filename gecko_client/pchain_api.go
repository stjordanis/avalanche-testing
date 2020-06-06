package gecko_client

import (
	"encoding/json"
	"github.com/palantir/stacktrace"
)

// TODO this file is quickly going to get unwieldy, so probably good to move it to its own package
const (
	pchainEndpoint = "ext/P"
)

type PChainApi struct {
	rpcRequester geckoJsonRpcRequester
}

type Validator struct {
	StartTime string `json:"startTime"`
	EndTime string	`json:"endTime"`
	StakeAmount string	`json:"stakeAmount"`
	Id string	`json:"id"`
}

type ValidatorList struct {
	Validators []Validator	`json:"validators"`
}

type GetValidatorsResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result ValidatorList	`json:"result"`
	Id int	`json:"id"`
}

func (api PChainApi) GetCurrentValidators() ([]Validator, error) {
	responseBodyBytes, err := api.rpcRequester.makeRpcRequest(pchainEndpoint, "platform.getCurrentValidators")
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error making request")
	}

	// TODO try moving this inside the MakeRequest method, even though Go doesn't have generics
	var response GetValidatorsResponse
	if err := json.Unmarshal(responseBodyBytes, &response); err != nil {
		return nil, stacktrace.Propagate(err, "Error unmarshalling JSON response")
	}
	return response.Result.Validators, nil
}
