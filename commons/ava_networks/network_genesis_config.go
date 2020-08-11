package ava_networks

/*
Encapsulates genesis block information about a chain
 */
type NetworkGenesisConfig struct {
	Stakers         []StakerIdentity
	FundedAddresses FundedAddress
}

/*
Encapsulates an already-funded address
 */
type FundedAddress struct {
	Address string
	PrivateKey string
}

/*
Represents a staker declared in the genesis config
 */
type StakerIdentity struct {
	NodeID string
	PrivateKey string
	TlsCert string
}
