package gecko_client

type Check struct {
	Message map[string]interface{} 	`json:"message"`
	Timestamp string	`json:"timestamp"`
	Duration int 	`json:"duration"`
	ContiguousFailures int 	`json:"contiguousFailures"`
	TimeOfFirstFailure string	`json:"timeOfFirstFailure"`
}

type LivenessInfo struct {
	Checks map[string]Check	`json:"checks"`
	Healthy bool	`json:"healthy"`
}

type GetLivenessResponse struct {
	JsonRpcVersion string       `json:"jsonrpc"`
	Result         LivenessInfo `json:"result"`
	Id             int          `json:"id"`
}
