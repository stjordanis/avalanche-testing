package gecko_client

type CreateUser struct {
	Success bool `json:"success"`
}

type CreateUserResponse struct {
	JsonRpcVersion string	`json:"jsonrpc"`
	Result CreateUser	`json:"result"`
	Id int	`json:"id"`
}