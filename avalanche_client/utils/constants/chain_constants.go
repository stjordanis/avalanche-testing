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

	XChainID, _ = ids.FromString("iWNpYgu4cCoxvSVywQy2egFCmgq9d9Tp1LeFURTVrHVZbmURK")
	PlatformChainID = ids.Empty
	CChainID, _ = ids.FromString("8XUYBzd61vw86MC2UmizjRdYFxcEKgQabZqx8hoSjzQiqoewQ")
	AvaxAssetID, _ = ids.FromString("o4TncKL4JSBuW1HJK19aE8UBdSXcNzFzF4RtPjqtawb5DPSrH")
}
