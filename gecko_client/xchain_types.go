package gecko_client

const TXN_ACCEPTED = "ACCEPTED"

type XChainTxnInfo struct {
	TxID string `json:"txID"`
}

type TxStatus struct {
	Status string `json:"status"`
}

type GetTxStatusResponse struct {
	JsonRpcVersion string                  `json:"jsonrpc"`
	Result         TxStatus `json:"result"`
	Id             int                     `json:"id"`
}

type UtxoIdInfo struct {
	TxID string `json:"txID"`
	OutputIndex int `json:"outputIndex"`
}

type AccountWithUtxoInfo struct {
	Balance string `json:"balance"`
	UtxoIDs []UtxoIdInfo `json:"utxoIDs"`
}

type GetBalanceResponse struct {
	JsonRpcVersion string                  `json:"jsonrpc"`
	Result         AccountWithUtxoInfo `json:"result"`
	Id             int                     `json:"id"`
}

type XChainExportAVAResponse struct {
	JsonRpcVersion string                  `json:"jsonrpc"`
	Result         XChainTxnInfo `json:"result"`
	Id             int                     `json:"id"`
}

type SendResponse struct {
	JsonRpcVersion string                  `json:"jsonrpc"`
	Result         XChainTxnInfo `json:"result"`
	Id             int                     `json:"id"`
}

type AddressInfo struct {
	Address string `json:"address"`
}

type CreateAddressResponse struct {
	JsonRpcVersion string                  `json:"jsonrpc"`
	Result         AddressInfo `json:"result"`
	Id             int                     `json:"id"`
}