package api

type Validator struct {
	StartTime string
	EndTime string
	StakeAmount string
	Id string
}

type ValidatorList []*Validator

type ValidatorResponse struct {
	Jsonrpc string
	Result map[string]ValidatorList
	Id int
}

func GetPChainEndpoint() string {
	return "ext/P"
}
