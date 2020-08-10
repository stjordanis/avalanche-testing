package gecko_client

import "github.com/ava-labs/gecko/api/health"

type Check struct {
	Message            map[string]interface{} `json:"message"`
	Timestamp          string                 `json:"timestamp"`
	Duration           int                    `json:"duration"`
	ContiguousFailures int                    `json:"contiguousFailures"`
	TimeOfFirstFailure string                 `json:"timeOfFirstFailure"`
}

type LivenessInfo struct {
	Checks  map[string]Check `json:"checks"`
	Healthy bool             `json:"healthy"`
}

type GetLivenessResponse struct {
	JsonRpcVersion string                  `json:"jsonrpc"`
	Result         health.GetLivenessReply `json:"result"`
	Id             int                     `json:"id"`
}
