// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/formatting"
	cjson "github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/vms/avm"

	"github.com/ava-labs/avalanche-testing/avalanche_client/utils"
)

type Client struct {
	requester utils.EndpointRequester
}

// Returns a Client for interacting with the X chain endpoint
func NewClient(uri, chain string, requestTimeout time.Duration) *Client {
	return &Client{
		requester: utils.NewEndpointRequester(uri, fmt.Sprintf("/ext/bc/%s", chain), "avm", requestTimeout),
	}
}

// IssueTx issues a transaction to a node and returns the TxID
func (c *Client) IssueTx(txBytes []byte) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("issueTx", &avm.FormattedTx{
		Tx: formatting.CB58{Bytes: txBytes},
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}

func (c *Client) GetTxStatus(txID ids.ID) (choices.Status, error) {
	res := &avm.GetTxStatusReply{}
	err := c.requester.SendRequest("getTxStatus", &api.JsonTxID{
		TxID: txID,
	}, res)
	if err != nil {
		return choices.Unknown, err
	}
	return res.Status, nil
}

func (c *Client) GetTx(txID ids.ID) ([]byte, error) {
	res := &avm.FormattedTx{}
	err := c.requester.SendRequest("getTx", &api.JsonTxID{
		TxID: txID,
	}, res)
	if err != nil {
		return nil, err
	}
	return res.Tx.Bytes, nil
}

// GetUTXOs returns the byte representation of the UTXOs controlled by [addrs]
func (c *Client) GetUTXOs(addrs []string, limit uint32, startAddress, startUTXOID string) (*avm.GetUTXOsReply, error) {
	res := &avm.GetUTXOsReply{}
	err := c.requester.SendRequest("getUTXOs", &avm.GetUTXOsArgs{
		Addresses: addrs,
		Limit:     cjson.Uint32(limit),
		StartIndex: avm.Index{
			Address: startAddress,
			UTXO:    startUTXOID,
		},
	}, res)
	return res, err
}

func (c *Client) GetAssetDescription(assetID string) (*avm.GetAssetDescriptionReply, error) {
	res := &avm.GetAssetDescriptionReply{}
	err := c.requester.SendRequest("getAssetDescription", &avm.GetAssetDescriptionArgs{
		AssetID: assetID,
	}, res)
	return res, err
}

func (c *Client) GetBalance(addr string, assetID string) (*avm.GetBalanceReply, error) {
	res := &avm.GetBalanceReply{}
	err := c.requester.SendRequest("getBalance", &avm.GetBalanceArgs{
		Address: addr,
		AssetID: assetID,
	}, res)
	return res, err
}

func (c *Client) GetAllBalances(addr string, assetID string) (*avm.GetAllBalancesReply, error) {
	res := &avm.GetAllBalancesReply{}
	err := c.requester.SendRequest("getAllBalances", &api.JsonAddress{
		Address: addr,
	}, res)
	return res, err
}

func (c *Client) CreateFixedCapAsset(
	user api.UserPass,
	name,
	symbol string,
	denomination byte,
	holders []*avm.Holder,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &avm.FormattedAssetID{}
	err := c.requester.SendRequest("createFixedCapAsset", &avm.CreateFixedCapAssetArgs{
		JsonSpendHeader: api.JsonSpendHeader{
			UserPass:       user,
			JsonFromAddrs:  api.JsonFromAddrs{From: from},
			JsonChangeAddr: api.JsonChangeAddr{ChangeAddr: changeAddr},
		},
		Name:           name,
		Symbol:         symbol,
		Denomination:   denomination,
		InitialHolders: holders,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.AssetID, nil
}

func (c *Client) CreateVariableCapAsset(
	user api.UserPass,
	name,
	symbol string,
	denomination byte,
	minters []avm.Owners,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &avm.FormattedAssetID{}
	err := c.requester.SendRequest("createVariableCapAsset", &avm.CreateVariableCapAssetArgs{
		JsonSpendHeader: api.JsonSpendHeader{
			UserPass:       user,
			JsonFromAddrs:  api.JsonFromAddrs{From: from},
			JsonChangeAddr: api.JsonChangeAddr{ChangeAddr: changeAddr},
		},
		Name:         name,
		Symbol:       symbol,
		Denomination: denomination,
		MinterSets:   minters,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.AssetID, nil
}

func (c *Client) CreateNFTAsset(
	user api.UserPass,
	name,
	symbol string,
	minters []avm.Owners,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &avm.FormattedAssetID{}
	err := c.requester.SendRequest("createNFTAsset", &avm.CreateNFTAssetArgs{
		JsonSpendHeader: api.JsonSpendHeader{
			UserPass:       user,
			JsonFromAddrs:  api.JsonFromAddrs{From: from},
			JsonChangeAddr: api.JsonChangeAddr{ChangeAddr: changeAddr},
		},
		Name:       name,
		Symbol:     symbol,
		MinterSets: minters,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.AssetID, nil
}

func (c *Client) CreateAddress(user api.UserPass) (string, error) {
	res := &api.JsonAddress{}
	err := c.requester.SendRequest("createAddress", &user, res)
	if err != nil {
		return "", err
	}
	return res.Address, nil
}

func (c *Client) ListAddresses(user api.UserPass) ([]string, error) {
	res := &api.JsonAddresses{}
	err := c.requester.SendRequest("listAddresses", &user, res)
	if err != nil {
		return nil, err
	}
	return res.Addresses, nil
}

func (c *Client) ExportKey(user api.UserPass, addr string) (string, error) {
	res := &avm.ExportKeyReply{}
	err := c.requester.SendRequest("exportKey", &avm.ExportKeyArgs{
		UserPass: user,
		Address:  addr,
	}, res)
	if err != nil {
		return "", err
	}
	return res.PrivateKey, nil
}

func (c *Client) ImportKey(user api.UserPass, privateKey string) (string, error) {
	res := &api.JsonAddress{}
	err := c.requester.SendRequest("importKey", &avm.ImportKeyArgs{
		UserPass:   user,
		PrivateKey: privateKey,
	}, res)
	if err != nil {
		return "", err
	}
	return res.Address, nil
}

func (c *Client) Send(
	user api.UserPass,
	amount uint64,
	assetID,
	to string,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("send", &avm.SendArgs{
		JsonSpendHeader: api.JsonSpendHeader{
			UserPass:       user,
			JsonFromAddrs:  api.JsonFromAddrs{From: from},
			JsonChangeAddr: api.JsonChangeAddr{ChangeAddr: changeAddr},
		},
		Amount:  cjson.Uint64(amount),
		AssetID: assetID,
		To:      to,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}

func (c *Client) Mint(
	user api.UserPass,
	amount uint64,
	assetID,
	to string,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("mint", &avm.MintArgs{
		JsonSpendHeader: api.JsonSpendHeader{
			UserPass:       user,
			JsonFromAddrs:  api.JsonFromAddrs{From: from},
			JsonChangeAddr: api.JsonChangeAddr{ChangeAddr: changeAddr},
		},
		Amount:  cjson.Uint64(amount),
		AssetID: assetID,
		To:      to,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}

func (c *Client) SendNFT(
	user api.UserPass,
	assetID string,
	groupID uint32,
	to string,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("sendNFT", &avm.SendNFTArgs{
		JsonSpendHeader: api.JsonSpendHeader{
			UserPass:       user,
			JsonFromAddrs:  api.JsonFromAddrs{From: from},
			JsonChangeAddr: api.JsonChangeAddr{ChangeAddr: changeAddr},
		},
		AssetID: assetID,
		GroupID: cjson.Uint32(groupID),
		To:      to,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}

func (c *Client) MintNFT(
	user api.UserPass,
	assetID string,
	payload []byte,
	to string,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("mintNFT", &avm.MintNFTArgs{
		JsonSpendHeader: api.JsonSpendHeader{
			UserPass:       user,
			JsonFromAddrs:  api.JsonFromAddrs{From: from},
			JsonChangeAddr: api.JsonChangeAddr{ChangeAddr: changeAddr},
		},
		AssetID: assetID,
		Payload: formatting.CB58{Bytes: payload},
		To:      to,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}

func (c *Client) ImportAVAX(user api.UserPass, to, sourceChain string) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("importAVAX", &avm.ImportAVAXArgs{
		UserPass:    user,
		To:          to,
		SourceChain: sourceChain,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}

func (c *Client) ExportAVAX(
	user api.UserPass,
	amount uint64,
	to string,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &api.JsonTxID{}
	err := c.requester.SendRequest("exportAVAX", &avm.ExportAVAXArgs{
		JsonSpendHeader: api.JsonSpendHeader{
			UserPass:       user,
			JsonFromAddrs:  api.JsonFromAddrs{From: from},
			JsonChangeAddr: api.JsonChangeAddr{ChangeAddr: changeAddr},
		},
		Amount: cjson.Uint64(amount),
		To:     to,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}
