package gecko_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestGetPeers(t *testing.T) {
	resultStr := `{
    "jsonrpc":"2.0",
    "id"     :1,
    "result" :{
        "peers":[
          {
             "ip":"206.189.137.87:9651",
             "publicIP":"206.189.137.87:9651",
             "id":"8PYXX47kqLDe2wD4oPbvRRchcnSzMA4J4",
             "version":"avalanche/0.5.0",
             "lastSent":"2020-06-01T15:23:02Z",
             "lastReceived":"2020-06-01T15:22:57Z"
          },
          {
             "ip":"158.255.67.151:9651",
             "publicIP":"158.255.67.151:9651",
             "id":"C14fr1n8EYNKyDfYixJ3rxSAVqTY3a8BP",
             "version":"avalanche/0.5.0",
             "lastSent":"2020-06-01T15:23:02Z",
             "lastReceived":"2020-06-01T15:22:34Z"
          },
          {
             "ip":"83.42.13.44:9651",
             "publicIP":"83.42.13.44:9651",
             "id":"LPbcSMGJ4yocxYxvS2kBJ6umWeeFbctYZ",
             "version":"avalanche/0.5.0",
             "lastSent":"2020-06-01T15:23:02Z",
             "lastReceived":"2020-06-01T15:22:55Z"
          }
        ]
    }
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	peers, err := client.InfoApi().GetPeers()
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, len(peers), 3)
}

func TestGetNodeId(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "nodeID": "5mb46qkSBj81k9g9e4VFjGGSbaaSLFRzD"
    },
    "id": 1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	nodeId, err := client.InfoApi().GetNodeId()
	assert.Nil(t, err, "Error message should be nil")
	assert.Equal(t, nodeId, "5mb46qkSBj81k9g9e4VFjGGSbaaSLFRzD")
}

func TestIsBootstrapped(t *testing.T) {
	resultStr := `{
    "jsonrpc": "2.0",
    "result": {
        "isBootstrapped": true
    },
    "id": 1
}`
	client := clientFromRequester(mockedJsonRpcRequester{resultStr: resultStr})
	isBootstrapped, err := client.InfoApi().IsBootstrapped("P")
	assert.Nil(t, err, "Error message should be nil")
	assert.True(t, isBootstrapped)
}
