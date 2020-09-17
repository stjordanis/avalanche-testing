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
	XChainID, _ = ids.FromString("v4hFSZTNNVdyomeMoXa77dAz4CdxU3cziSb45TB7mfXUmy7C7")
	PlatformChainID = ids.Empty
	CChainID, _ = ids.FromString("2m6aMgMBJWsmT4Hv448n6sNAwGMFfugBvdU6PdY5oxZge4qb1W")
	AvaxAssetID, _ = ids.FromString("SSUAMrVdqYuvybAMGNitTYSAnE4T5fVdVDB82ped1qQ9f8DDM")
}
