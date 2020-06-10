package gecko_client

// ============= Blockchain ====================
type BlockchainCreationInfo struct {
	UnsignedTx string 	`json:"unsignedTx"`
}

type CreateBlockchainResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result BlockchainCreationInfo	`json:"result"`
	Id int	`json:"id"`
}

type BlockchainStatus struct {
	Status string	`json:"status"`
}

type GetBlockchainStatusResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result BlockchainStatus	`json:"result"`
	Id int	`json:"id"`
}

// ============= Accounts ====================
type AccountInfo struct {
	Address string 	`json:"address"`
}

type CreateAccountResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result AccountInfo	`json:"result"`
	Id int	`json:"id"`
}

type ImportKeyResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result AccountInfo	`json:"result"`
	Id int	`json:"id"`
}

type PrivateKeyInfo struct {
	PrivateKey string	`json:"privateKey"`
}

type ExportKeyResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result PrivateKeyInfo	`json:"result"`
	Id int	`json:"id"`
}

// ============= Validators ====================
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

