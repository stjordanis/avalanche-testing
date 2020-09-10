package networks

// DefaultLocalNetGenesisConfig contains the private keys and node IDs that come from Gecko for the 5 bootstrapper nodes.
// When using Gecko with the 'local' testnet option, the P-chain comes preloaded with five bootstrapper nodes whose node
// IDs are hardcoded in Gecko source. Node IDs are determined based off the TLS keys of the nodes, so to ensure that
// we can launch nodes with the same node ID (to validate, else we wouldn't be able to validate at all), the Gecko
// source code also provides the private keys for these nodes.
var DefaultLocalNetGenesisConfig = NetworkGenesisConfig{
	Stakers: defaultStakers,
	// hardcoded in Gecko in "genesis/config.go". needed to distribute genesis funds in tests
	FundedAddresses: FundedAddress{
		"6Y3kysjF9jnHnYkdS9yGAuoHyae2eNmeV",
		/*
			 	It's okay to have privateKey here because its a hardcoded value available in the Gecko codebase.
				It is necessary to have this privateKey in order to transfer funds to test accounts in the test net.
				This privateKey only applies to local test nets, it has nothing to do with the public test net or main net.
		*/
		"PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN",
	},
}

/*
In Gecko, you need at least $snow_consensus stakers for anything to happen. But, you can't register new stakers... without
$snow_conensus stakers already staking. Thus, you have to start with some staker IDs already registered. To do this, Gecko
hardcodes 5 staker IDs already registered on the PChain:
https://github.com/ava-labs/avalanche-go/blob/master/genesis/config.go#L407

These IDs are those stakers, and all local testnets
*/
var defaultStakers = []StakerIdentity{
	staker1,
	staker2,
	staker3,
	staker4,
	staker5,
}

var staker1 = StakerIdentity{
	Staker1NodeID,
	Staker1PrivateKey,
	Staker1Cert,
}

var staker2 = StakerIdentity{
	Staker2NodeID,
	Staker2PrivateKey,
	Staker2Cert,
}

var staker3 = StakerIdentity{
	Staker3NodeID,
	Staker3PrivateKey,
	Staker3Cert,
}

var staker4 = StakerIdentity{
	Staker4NodeID,
	Staker4PrivateKey,
	Staker4Cert,
}

var staker5 = StakerIdentity{
	Staker5NodeID,
	Staker5PrivateKey,
	Staker5Cert,
}
