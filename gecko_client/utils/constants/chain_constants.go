package constants

import (
	"github.com/ava-labs/gecko/ids"
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
	XChainID, _ = ids.FromString("2VvmkRw4yrz8tPrVnCCbvEK1JxNyujpqhmU6SGonxMpkWBx9UD")
	PlatformChainID = ids.Empty
	CChainID, _ = ids.FromString("f5DjTrC9YJPagt9ogKgKPYpp7KMaCBKsv7AeqfonpTiw6rBec")
	AvaxAssetID, _ = ids.FromString("2TrXx5kLGWa9RP3RiYWi7VkmNbppwPU4DCmTdqwuKzGFE7fsvP")
}
