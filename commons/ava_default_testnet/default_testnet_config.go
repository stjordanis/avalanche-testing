package ava_default_testnet

/*
	For the default testnet, there are five bootstrappers. They have hardcoded bootstrapperIDs that correspond to their TLS certs.
	This must be hardcoded because Gecko requires specifying the bootstrapperID
	along with the bootstrapperIP when connecting to bootstrappers in TLS mode.

	The hardcoded IDs in this file are the known Gecko ID for a node using the private key and cert in this file.
	They are also hardcoded in the gecko source code as the IDs for the initial stakers of the default testnet.
*/

type TestNet struct {
	DefaultStakers []StakerIdentity
	FundedAddresses FundedAddress
}

type FundedAddress struct {
	Address string
	PrivateKey string
}

type StakerIdentity struct {
	NodeID string
	PrivateKey string
	TlsCert string
}

var DefaultTestNet = TestNet{
	DefaultStakers: defaultStakers,
	// hardcoded in Gecko in "genesis/config.go". needed to distribute genesis funds in tests
	FundedAddresses: FundedAddress{
		"6Y3kysjF9jnHnYkdS9yGAuoHyae2eNmeV",
		"ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN",
	},
}

var defaultStakers = []StakerIdentity{
	staker1,
	staker2,
	staker3,
	staker4,
	staker5,
}

var staker1 = StakerIdentity{
	STAKER_1_NODE_ID,
	STAKER_1_PRIVATE_KEY,
	STAKER_1_CERT,
}

var staker2 = StakerIdentity{
	STAKER_2_NODE_ID,
	STAKER_2_PRIVATE_KEY,
	STAKER_2_CERT,
}

var staker3 = StakerIdentity{
	STAKER_3_NODE_ID,
	STAKER_3_PRIVATE_KEY,
	STAKER_3_CERT,
}

var staker4 = StakerIdentity{
	STAKER_4_NODE_ID,
	STAKER_4_PRIVATE_KEY,
	STAKER_4_CERT,
}

var staker5 = StakerIdentity{
	STAKER_5_NODE_ID,
	STAKER_5_PRIVATE_KEY,
	STAKER_5_CERT,
}
