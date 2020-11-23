// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/vms/avm/vmargs"

	"github.com/ava-labs/avalanchego/api/apiargs"

	"github.com/ava-labs/avalanche-testing/avalanche_client/utils"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/utils/formatting"
	cjson "github.com/ava-labs/avalanchego/utils/json"
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
	res := &apiargs.JSONTxID{}

	txEncodedStr, err := formatting.Encode(formatting.Hex, txBytes)
	if err != nil {
		return [32]byte{}, err
	}

	err = c.requester.SendRequest("issueTx", &apiargs.FormattedTx{
		Tx:       txEncodedStr,
		Encoding: formatting.Hex,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}

func (c *Client) GetTxStatus(txID ids.ID) (choices.Status, error) {
	res := &vmargs.GetTxStatusReply{}
	err := c.requester.SendRequest("getTxStatus", &apiargs.JSONTxID{
		TxID: txID,
	}, res)
	if err != nil {
		return choices.Unknown, err
	}
	return res.Status, nil
}

func (c *Client) GetTx(txID ids.ID) ([]byte, error) {
	res := &apiargs.FormattedTx{}
	err := c.requester.SendRequest("getTx", &apiargs.GetTxArgs{
		TxID:     txID,
		Encoding: formatting.Hex,
	}, res)
	if err != nil {
		return nil, err
	}

	formatted, err := formatting.Decode(formatting.Hex, res.Tx)
	if err != nil {
		return nil, err
	}
	return formatted, nil
}

// GetUTXOs returns the byte representation of the UTXOs controlled by [addrs]
func (c *Client) GetUTXOs(addrs []string, limit uint32, startAddress, startUTXOID string) ([][]byte, vmargs.Index, error) {
	res := &vmargs.GetUTXOsReply{}
	err := c.requester.SendRequest("getUTXOs", &vmargs.GetUTXOsArgs{
		Addresses: addrs,
		Limit:     cjson.Uint32(limit),
		StartIndex: vmargs.Index{
			Address: startAddress,
			UTXO:    startUTXOID,
		},
		Encoding: formatting.Hex,
	}, res)
	if err != nil {
		return nil, vmargs.Index{}, err
	}

	utxos := make([][]byte, len(res.UTXOs))
	for i, utxo := range res.UTXOs {
		formatted, err := formatting.Decode(formatting.Hex, utxo)
		if err != nil {
			return nil, vmargs.Index{}, err
		}
		utxos[i] = formatted
	}
	return utxos, res.EndIndex, nil
}

func (c *Client) GetAssetDescription(assetID string) (*vmargs.GetAssetDescriptionReply, error) {
	res := &vmargs.GetAssetDescriptionReply{}
	err := c.requester.SendRequest("getAssetDescription", &vmargs.GetAssetDescriptionArgs{
		AssetID: assetID,
	}, res)
	return res, err
}

func (c *Client) GetBalance(addr string, assetID string) (*vmargs.GetBalanceReply, error) {
	res := &vmargs.GetBalanceReply{}
	err := c.requester.SendRequest("getBalance", &vmargs.GetBalanceArgs{
		Address: addr,
		AssetID: assetID,
	}, res)
	return res, err
}

func (c *Client) GetAllBalances(addr string, assetID string) (*vmargs.GetAllBalancesReply, error) {
	res := &vmargs.GetAllBalancesReply{}
	err := c.requester.SendRequest("getAllBalances", &apiargs.JSONAddress{
		Address: addr,
	}, res)
	return res, err
}

func (c *Client) CreateFixedCapAsset(
	user apiargs.UserPass,
	name,
	symbol string,
	denomination byte,
	holders []*vmargs.Holder,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &vmargs.FormattedAssetID{}
	err := c.requester.SendRequest("createFixedCapAsset", &vmargs.CreateFixedCapAssetArgs{
		JSONSpendHeader: apiargs.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  apiargs.JSONFromAddrs{From: from},
			JSONChangeAddr: apiargs.JSONChangeAddr{ChangeAddr: changeAddr},
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
	user apiargs.UserPass,
	name,
	symbol string,
	denomination byte,
	minters []vmargs.Owners,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &vmargs.FormattedAssetID{}
	err := c.requester.SendRequest("createVariableCapAsset", &vmargs.CreateVariableCapAssetArgs{
		JSONSpendHeader: apiargs.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  apiargs.JSONFromAddrs{From: from},
			JSONChangeAddr: apiargs.JSONChangeAddr{ChangeAddr: changeAddr},
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
	user apiargs.UserPass,
	name,
	symbol string,
	minters []vmargs.Owners,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &vmargs.FormattedAssetID{}
	err := c.requester.SendRequest("createNFTAsset", &vmargs.CreateNFTAssetArgs{
		JSONSpendHeader: apiargs.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  apiargs.JSONFromAddrs{From: from},
			JSONChangeAddr: apiargs.JSONChangeAddr{ChangeAddr: changeAddr},
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

func (c *Client) CreateAddress(user apiargs.UserPass) (string, error) {
	res := &apiargs.JSONAddress{}
	err := c.requester.SendRequest("createAddress", &user, res)
	if err != nil {
		return "", err
	}
	return res.Address, nil
}

func (c *Client) ListAddresses(user apiargs.UserPass) ([]string, error) {
	res := &apiargs.JSONAddresses{}
	err := c.requester.SendRequest("listAddresses", &user, res)
	if err != nil {
		return nil, err
	}
	return res.Addresses, nil
}

func (c *Client) ExportKey(user apiargs.UserPass, addr string) (string, error) {
	res := &vmargs.ExportKeyReply{}
	err := c.requester.SendRequest("exportKey", &vmargs.ExportKeyArgs{
		UserPass: user,
		Address:  addr,
	}, res)
	if err != nil {
		return "", err
	}
	return res.PrivateKey, nil
}

func (c *Client) ImportKey(user apiargs.UserPass, privateKey string) (string, error) {
	res := &apiargs.JSONAddress{}
	err := c.requester.SendRequest("importKey", &vmargs.ImportKeyArgs{
		UserPass:   user,
		PrivateKey: privateKey,
	}, res)
	if err != nil {
		return "", err
	}
	return res.Address, nil
}

func (c *Client) Send(
	user apiargs.UserPass,
	amount uint64,
	assetID,
	to string,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &apiargs.JSONTxID{}
	err := c.requester.SendRequest("send", &vmargs.SendArgs{
		JSONSpendHeader: apiargs.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  apiargs.JSONFromAddrs{From: from},
			JSONChangeAddr: apiargs.JSONChangeAddr{ChangeAddr: changeAddr},
		},
		SendOutput: vmargs.SendOutput{
			Amount:  cjson.Uint64(amount),
			AssetID: assetID,
			To:      to,
		},
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}

func (c *Client) Mint(
	user apiargs.UserPass,
	amount uint64,
	assetID,
	to string,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &apiargs.JSONTxID{}
	err := c.requester.SendRequest("mint", &vmargs.MintArgs{
		JSONSpendHeader: apiargs.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  apiargs.JSONFromAddrs{From: from},
			JSONChangeAddr: apiargs.JSONChangeAddr{ChangeAddr: changeAddr},
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
	user apiargs.UserPass,
	assetID string,
	groupID uint32,
	to string,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &apiargs.JSONTxID{}
	err := c.requester.SendRequest("sendNFT", &vmargs.SendNFTArgs{
		JSONSpendHeader: apiargs.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  apiargs.JSONFromAddrs{From: from},
			JSONChangeAddr: apiargs.JSONChangeAddr{ChangeAddr: changeAddr},
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
	user apiargs.UserPass,
	assetID string,
	payload []byte,
	to string,
	from []string,
	changeAddr string,
) (ids.ID, error) {

	payloadStr, err := formatting.Encode(formatting.Hex, payload)
	if err != nil {
		return [32]byte{}, err
	}

	res := &apiargs.JSONTxID{}
	err = c.requester.SendRequest("mintNFT", &vmargs.MintNFTArgs{
		JSONSpendHeader: apiargs.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  apiargs.JSONFromAddrs{From: from},
			JSONChangeAddr: apiargs.JSONChangeAddr{ChangeAddr: changeAddr},
		},
		AssetID:  assetID,
		Payload:  payloadStr,
		Encoding: formatting.Hex,
		To:       to,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}

func (c *Client) ImportAVAX(user apiargs.UserPass, to, sourceChain string) (ids.ID, error) {
	res := &apiargs.JSONTxID{}
	err := c.requester.SendRequest("importAVAX", &vmargs.ImportArgs{
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
	user apiargs.UserPass,
	amount uint64,
	to string,
	from []string,
	changeAddr string,
) (ids.ID, error) {
	res := &apiargs.JSONTxID{}
	err := c.requester.SendRequest("exportAVAX", &vmargs.ExportAVAXArgs{
		JSONSpendHeader: apiargs.JSONSpendHeader{
			UserPass:       user,
			JSONFromAddrs:  apiargs.JSONFromAddrs{From: from},
			JSONChangeAddr: apiargs.JSONChangeAddr{ChangeAddr: changeAddr},
		},
		Amount: cjson.Uint64(amount),
		To:     to,
	}, res)
	if err != nil {
		return ids.Empty, err
	}
	return res.TxID, nil
}
