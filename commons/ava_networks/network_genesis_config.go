package ava_networks

type NetworkGenesisConfig struct {
	Stakers         []StakerIdentity
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
