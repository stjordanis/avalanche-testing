package gecko_client

type NodeID struct {
	NodeID string `json:"nodeID"`
}

type Peer struct {
	IP string	`json:"ip"`
	PublicIP string 	`json:"publicIP"`
	Id string	`json:"id"`
	Version string	`json:"version"`
	LastSent string 	`json:"lastSent"`
	LastReceived string	`json:"lastReceived"`
}

type PeerList struct {
	Peers []Peer	`json:"peers"`
}

type GetPeersResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result PeerList	`json:"result"`
	Id int	`json:"id"`
}

type GetNodeIDResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result NodeID	`json:"result"`
	Id int	`json:"id"`
}
