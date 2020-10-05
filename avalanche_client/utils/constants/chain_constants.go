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
	CChainID, _ = ids.FromString("WKNkfmNxgqpKPe9Q12UCoTuGYXX5JbQn2tf2WTpNTJeQrezqa")
	AvaxAssetID, _ = ids.FromString("2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe")
}
