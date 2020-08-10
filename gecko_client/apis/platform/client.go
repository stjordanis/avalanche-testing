// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platform

import (
	"time"

	"github.com/ava-labs/gecko/api"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/utils/formatting"
	cjson "github.com/ava-labs/gecko/utils/json"
	"github.com/ava-labs/gecko/vms/platformvm"

	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/utils"
)

type Client struct {
	requester utils.EndpointRequester
}

// Returns a Client for interacting with the P Chain endpoint
func NewClient(uri string, requestTimeout time.Duration) *Client {
	return &Client{
		requester: utils.NewEndpointRequester(uri, "/ext/P", "platform", requestTimeout),
	}
}

func (c *Client) GetHeight() (uint64, error) {
	res := &platformvm.GetHeightResponse{}
	err := c.requester.SendRequest("getHeight", struct{}{}, res)
	return uint64(res.Height), err
}

func (c *Client) ExportKey(user api.UserPass, address string) (string, error) {
	res := &platformvm.ExportKeyReply{}
	err := c.requester.SendRequest("exportKey", &platformvm.ExportKeyArgs{
		UserPass: user,
		Address:  address,
	}, res)
	return res.PrivateKey, err
}

func (c *Client) ImportKey(user api.UserPass, privateKey string) (string, error) {
	res := &api.JsonAddress{}
	err := c.requester.SendRequest("importKey", &platformvm.ImportKeyArgs{
		UserPass:   user,
		PrivateKey: privateKey,
	}, res)
	return res.Address, err
}

func (c *Client) GetBalance(address string) (*platformvm.GetBalanceResponse, error) {
	res := &platformvm.GetBalanceResponse{}
	err := c.requester.SendRequest("getBalance", &platformvm.GetBalanceArgs{
		Address: address,
	}, res)
	return res, err
}

func (c *Client) CreateAddress(user api.UserPass) (string, error) {
	res := &api.JsonAddress{}
	err := c.requester.SendRequest("createAddress", &user, res)
	return res.Address, err
}

func (c *Client) ListAddresses(user api.UserPass) ([]string, error) {
	res := &api.JsonAddresses{}
	err := c.requester.SendRequest("listAddresses", &user, res)
	return res.Addresses, err
}

func (c *Client) GetUTXOs(addresses []string) ([][]byte, error) {
	res := &platformvm.GetUTXOsResponse{}
	err := c.requester.SendRequest("getUTXOs", &platformvm.GetUTXOsArgs{
		Addresses: addresses,
	}, res)
	if err != nil {
		return nil, err
	}
	utxos := make([][]byte, len(res.UTXOs))
	for i, utxo := range res.UTXOs {
		utxos[i] = utxo.Bytes
	}
	return utxos, nil
}

func (c *Client) GetSubnets(ids []ids.ID) ([]platformvm.APISubnet, error) {
	res := &platformvm.GetSubnetsResponse{}
	err := c.requester.SendRequest("getSubnets", &platformvm.GetSubnetsArgs{
		IDs: ids,
	}, res)
	return res.Subnets, err
}

func (c *Client) GetCurrentValidators(subnetID ids.ID) ([]platformvm.FormattedAPIValidator, error) {
	res := &platformvm.GetCurrentValidatorsReply{}
	err := c.requester.SendRequest("getCurrentValidators", &platformvm.GetCurrentValidatorsArgs{
		SubnetID: subnetID,
	}, res)
	return res.Validators, err
}

func (c *Client) GetPendingValidators(subnetID ids.ID) ([]platformvm.FormattedAPIValidator, error) {
	res := &platformvm.GetPendingValidatorsReply{}
	err := c.requester.SendRequest("getPendingValidators", &platformvm.GetPendingValidatorsArgs{
		SubnetID: subnetID,
	}, res)
	return res.Validators, err
}

// TODO standardize the return format along with GetCurrentValidators and GetPendingValidators
func (c *Client) SampleValidators(subnetID ids.ID, sampleSize uint16) (*platformvm.SampleValidatorsReply, error) {
	res := &platformvm.SampleValidatorsReply{}
	err := c.requester.SendRequest("sampleValidators", &platformvm.SampleValidatorsArgs{
		SubnetID: subnetID,
		Size:     cjson.Uint16(sampleSize),
	}, res)
	return res, err
}

// AddDefaultSubnetValidator...
func (c *Client) AddDefaultSubnetValidator(user api.UserPass, destination, nodeID string, stakeAmount, startTime, endTime uint64, delegationFeeRate uint32) (ids.ID, error) {
	res := &api.JsonTxID{}
	jsonStakeAmount := cjson.Uint64(stakeAmount)
	err := c.requester.SendRequest("addDefaultSubnetValidator", &platformvm.AddDefaultSubnetValidatorArgs{
		UserPass: user,
		FormattedAPIDefaultSubnetValidator: platformvm.FormattedAPIDefaultSubnetValidator{
			Destination:       destination,
			DelegationFeeRate: cjson.Uint32(delegationFeeRate),
			FormattedAPIValidator: platformvm.FormattedAPIValidator{
				ID:          nodeID,
				StakeAmount: &jsonStakeAmount,
				StartTime:   cjson.Uint64(startTime),
				EndTime:     cjson.Uint64(endTime),
			},
		},
	}, res)
	return res.TxID, err
}

// AddDefaultSubnetDelegator...
func (c *Client) AddDefaultSubnetDelegator(user api.UserPass, destination, nodeID string, stakeAmount, startTime, endTime uint64) (ids.ID, error) {
	res := &api.JsonTxID{}
	jsonStakeAmount := cjson.Uint64(stakeAmount)
	err := c.requester.SendRequest("addDefaultSubnetDelegator", &platformvm.AddDefaultSubnetDelegatorArgs{
		UserPass: user,
		FormattedAPIValidator: platformvm.FormattedAPIValidator{
			ID:          nodeID,
			StakeAmount: &jsonStakeAmount,
			StartTime:   cjson.Uint64(startTime),
			EndTime:     cjson.Uint64(endTime),
		},
		Destination: destination,
	}, res)
	return res.TxID, err
}

// AddNonDefaultSubnetValidator...
func (c *Client) AddNonDefaultSubnetValidator(user api.UserPass, destination, nodeID string, stakeAmount, startTime, endTime uint64, subnetID string) (ids.ID, error) {
	res := &api.JsonTxID{}
	jsonStakeAmount := cjson.Uint64(stakeAmount)
	err := c.requester.SendRequest("addNonDefaultSubnetValidator", &platformvm.AddNonDefaultSubnetValidatorArgs{
		UserPass: user,
		FormattedAPIValidator: platformvm.FormattedAPIValidator{
			ID:          nodeID,
			StakeAmount: &jsonStakeAmount,
			StartTime:   cjson.Uint64(startTime),
			EndTime:     cjson.Uint64(endTime),
		},
		SubnetID: subnetID,
	}, res)
	return res.TxID, err
}

func (c *Client) CreateSubnet(user api.UserPass, subnet platformvm.APISubnet) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("createSubnet", &platformvm.CreateSubnetArgs{
		UserPass:  user,
		APISubnet: subnet,
	}, res)
	return res.TxID, err
}

func (c *Client) ExportAVA(user api.UserPass, to string, amount uint64) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("exportAVA", &platformvm.ExportAVAArgs{
		UserPass: user,
		To:       to,
		Amount:   cjson.Uint64(amount),
	}, res)
	return res.TxID, err
}

func (c *Client) ImportAVA(user api.UserPass, to string) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("importAVA", &platformvm.ImportAVAArgs{
		UserPass: user,
		To:       to,
	}, res)
	return res.TxID, err
}

func (c *Client) CreateBlockchain(user api.UserPass, subnetID ids.ID, vmID string, fxIDs []string, name string, genesisData []byte) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("createBlockchain", &platformvm.CreateBlockchainArgs{
		UserPass:    user,
		SubnetID:    subnetID,
		VMID:        vmID,
		FxIDs:       fxIDs,
		Name:        name,
		GenesisData: formatting.CB58{Bytes: genesisData},
	}, res)
	return res.TxID, err
}

func (c *Client) GetBlockchainStatus(blockchainID string) (platformvm.Status, error) {
	res := &platformvm.GetBlockchainStatusReply{}
	err := c.requester.SendRequest("getBlockchainStatus", &platformvm.GetBlockchainStatusArgs{
		BlockchainID: blockchainID,
	}, res)
	return res.Status, err
}

// Returns the ID of the Subnet that validates [blockchainID]
func (c *Client) ValidatedBy(blockchainID ids.ID) (ids.ID, error) {
	res := &platformvm.ValidatedByResponse{}
	err := c.requester.SendRequest("validatedBy", &platformvm.ValidatedByArgs{
		BlockchainID: blockchainID,
	}, res)
	return res.SubnetID, err
}

// Returns the list of blockchains that are validated by the subnet with ID [subnetID]
func (c *Client) Validates(subnetID ids.ID) ([]ids.ID, error) {
	res := &platformvm.ValidatesResponse{}
	err := c.requester.SendRequest("validates", &platformvm.ValidatesArgs{
		SubnetID: subnetID,
	}, res)
	return res.BlockchainIDs, err
}

// Returns the list of blockchains on the platform
func (c *Client) GetBlockchains() ([]platformvm.APIBlockchain, error) {
	res := &platformvm.GetBlockchainsResponse{}
	err := c.requester.SendRequest("getBlockchains", struct{}{}, res)
	return res.Blockchains, err
}

func (c *Client) GetTx(txID ids.ID) ([]byte, error) {
	res := &platformvm.GetTxResponse{}
	err := c.requester.SendRequest("getTx", &platformvm.GetTxArgs{
		TxID: txID,
	}, res)
	return res.Tx.Bytes, err
}

func (c *Client) GetTxStatus(txID ids.ID) (platformvm.Status, error) {
	res := new(platformvm.Status)
	err := c.requester.SendRequest("getTxStatus", &platformvm.GetTxStatusArgs{
		TxID: txID,
	}, res)
	return *res, err
}
