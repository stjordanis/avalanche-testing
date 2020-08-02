package ava_networks



/*
When using Gecko with the 'local' testnet option, the P-chain comes preloaded with five bootstrapper nodes whose node
	IDs are hardcoded in Gecko source. Node IDs are determined based off the TLS keys of the nodes, so to ensure that
	we can launch nodes with the same node ID (to validate, else we wouldn't be able to validate at all), the Gecko
	source code also provides the private keys for these nodes.

This struct contains the private keys and node IDs that come from Gecko for the 5 bootstrapper nodes.
*/
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
		"ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN",
	},
}

/*
In Gecko, you need at least $snow_consensus stakers for anything to happen. But, you can't register new stakers... without
$snow_conensus stakers already staking. Thus, you have to start with some staker IDs already registered. To do this, Gecko
hardcodes 5 staker IDs already registered on the PChain:
https://github.com/ava-labs/gecko/blob/master/genesis/config.go#L407

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
