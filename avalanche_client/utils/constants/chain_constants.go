package constants

import (
	"github.com/ava-labs/avalanchego/ids"
)

var (
	// XChainID ...
	XChainID ids.ID
	// PlatformChainID ...
	PlatformChainID ids.ID
	// CChainID ...
	CChainID ids.ID
	// AvaxAssetID ...
	AvaxAssetID ids.ID
)

func init() {
	XChainID, _ = ids.FromString("2eNy1mUFdmaxXNj1eQHUe7Np4gju9sJsEtWQ4MX3ToiNKuADed")
	PlatformChainID = ids.Empty
	CChainID, _ = ids.FromString("2irUG9d7xeZbMYWLWo97Uv2oT9BrZfA4v5J28YJTeS6oeq4sBj")
	AvaxAssetID, _ = ids.FromString("2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe")
}
