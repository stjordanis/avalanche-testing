// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package keystore

import (
	"time"

	"github.com/ava-labs/avalanche-testing/gecko_client/utils"
	"github.com/ava-labs/gecko/utils/formatting"

	"github.com/ava-labs/gecko/api"
	"github.com/ava-labs/gecko/api/keystore"
)

type Client struct {
	requester utils.EndpointRequester
}

func NewClient(uri string, requestTimeout time.Duration) *Client {
	return &Client{
		requester: utils.NewEndpointRequester(uri, "/ext/keystore", "keystore", requestTimeout),
	}
}

func (c *Client) CreateUser(user api.UserPass) (bool, error) {
	res := &api.SuccessResponse{}
	err := c.requester.SendRequest("createUser", &user, res)
	return res.Success, err
}

func (c *Client) ListUsers() ([]string, error) {
	res := &keystore.ListUsersReply{}
	err := c.requester.SendRequest("listUsers", struct{}{}, res)
	return res.Users, err
}

func (c *Client) ExportUser(user api.UserPass) ([]byte, error) {
	res := &keystore.ExportUserReply{}
	err := c.requester.SendRequest("exportUser", &user, res)
	return res.User.Bytes, err
}

func (c *Client) ImportUser(user api.UserPass, account []byte) (bool, error) {
	res := &api.SuccessResponse{}
	err := c.requester.SendRequest("importUser", &keystore.ImportUserArgs{
		UserPass: user,
		User:     formatting.HexWrapper{Bytes: account},
	}, res)
	return res.Success, err
}

func (c *Client) DeleteUser(user api.UserPass) (bool, error) {
	res := &api.SuccessResponse{}
	err := c.requester.SendRequest("deleteUser", &user, res)
	return res.Success, err
}
