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
	XChainID, _ = ids.FromString("2G6XnaMqqFj6tkYPw8i3eHFoHnQqfo2yaS5BAtNEL2Knayq6qP")
	PlatformChainID = ids.Empty
	CChainID, _ = ids.FromString("2LcZK7Cp7LQNbvEbPfFvpr5qJqUjodghDndbWDZ7e6KYGMQ4jG")
	AvaxAssetID, _ = ids.FromString("2PQ2p3uTKfpS8Kmpss76NtDxdzikZw7uQj4x4ZKYHNRqWu5fMj")
}
