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
<<<<<<< HEAD
	XChainID, _ = ids.FromString("2eNy1mUFdmaxXNj1eQHUe7Np4gju9sJsEtWQ4MX3ToiNKuADed")
	PlatformChainID = ids.Empty
	CChainID, _ = ids.FromString("WKNkfmNxgqpKPe9Q12UCoTuGYXX5JbQn2tf2WTpNTJeQrezqa")
	AvaxAssetID, _ = ids.FromString("2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe")
=======
	XChainID, _ = ids.FromString("iWNpYgu4cCoxvSVywQy2egFCmgq9d9Tp1LeFURTVrHVZbmURK")
	PlatformChainID = ids.Empty
	CChainID, _ = ids.FromString("8XUYBzd61vw86MC2UmizjRdYFxcEKgQabZqx8hoSjzQiqoewQ")
	AvaxAssetID, _ = ids.FromString("o4TncKL4JSBuW1HJK19aE8UBdSXcNzFzF4RtPjqtawb5DPSrH")
>>>>>>> Subject:
}
