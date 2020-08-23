package admin

import (
	"testing"

	"github.com/ava-labs/avalanche-testing/gecko_client/apis/test"
	"github.com/ava-labs/avalanche-testing/gecko_client/utils"
	"github.com/ava-labs/gecko/api"
	"github.com/ava-labs/gecko/api/admin"
)

type mockClient struct {
	response interface{}
	err      error
}

// NewMockClient returns a mock client for testing
func NewMockClient(response interface{}, err error) utils.EndpointRequester {
	return &mockClient{
		response: response,
		err:      err,
	}
}

func (mc *mockClient) SendRequest(method string, params interface{}, reply interface{}) error {
	if mc.err != nil {
		return mc.err
	}

	switch p := reply.(type) {
	case *api.SuccessResponse:
		response := mc.response.(api.SuccessResponse)
		*p = response
	case *admin.StacktraceReply:
		response := mc.response.(admin.StacktraceReply)
		*p = response
	default:
		panic("illegal type")
	}
	return nil
}

func TestStartCPUProfiler(t *testing.T) {
	tests := test.GetSuccessResponseTests()

	for _, test := range tests {
		mockClient := Client{requester: NewMockClient(api.SuccessResponse{Success: test.Success}, test.Err)}
		success, err := mockClient.StartCPUProfiler()
		// if there is error as expected, the test passes
		if err != nil && test.Err != nil {
			continue
		}
		if err != nil {
			t.Fatalf("Unexepcted error: %s", err)
		}
		if success != test.Success {
			t.Fatalf("Expected success response to be: %v, but found: %v", test.Success, success)
		}
	}
}

func TestStopCPUProfiler(t *testing.T) {
	tests := test.GetSuccessResponseTests()

	for _, test := range tests {
		mockClient := Client{requester: NewMockClient(api.SuccessResponse{Success: test.Success}, test.Err)}
		success, err := mockClient.StopCPUProfiler()
		// if there is error as expected, the test passes
		if err != nil && test.Err != nil {
			continue
		}
		if err != nil {
			t.Fatalf("Unexepcted error: %s", err)
		}
		if success != test.Success {
			t.Fatalf("Expected success response to be: %v, but found: %v", test.Success, success)
		}
	}
}

func TestMemoryProfile(t *testing.T) {
	tests := test.GetSuccessResponseTests()

	for _, test := range tests {
		mockClient := Client{requester: NewMockClient(api.SuccessResponse{Success: test.Success}, test.Err)}
		success, err := mockClient.MemoryProfile()
		// if there is error as expected, the test passes
		if err != nil && test.Err != nil {
			continue
		}
		if err != nil {
			t.Fatalf("Unexepcted error: %s", err)
		}
		if success != test.Success {
			t.Fatalf("Expected success response to be: %v, but found: %v", test.Success, success)
		}
	}
}

func TestLockProfile(t *testing.T) {
	tests := test.GetSuccessResponseTests()

	for _, test := range tests {
		mockClient := Client{requester: NewMockClient(api.SuccessResponse{Success: test.Success}, test.Err)}
		success, err := mockClient.LockProfile()
		// if there is error as expected, the test passes
		if err != nil && test.Err != nil {
			continue
		}
		if err != nil {
			t.Fatalf("Unexepcted error: %s", err)
		}
		if success != test.Success {
			t.Fatalf("Expected success response to be: %v, but found: %v", test.Success, success)
		}
	}
}

func TestAlias(t *testing.T) {
	tests := test.GetSuccessResponseTests()

	for _, test := range tests {
		mockClient := Client{requester: NewMockClient(api.SuccessResponse{Success: test.Success}, test.Err)}
		success, err := mockClient.Alias("alias", "alias2")
		// if there is error as expected, the test passes
		if err != nil && test.Err != nil {
			continue
		}
		if err != nil {
			t.Fatalf("Unexepcted error: %s", err)
		}
		if success != test.Success {
			t.Fatalf("Expected success response to be: %v, but found: %v", test.Success, success)
		}
	}
}

func TestAliasChain(t *testing.T) {
	tests := test.GetSuccessResponseTests()

	for _, test := range tests {
		mockClient := Client{requester: NewMockClient(api.SuccessResponse{Success: test.Success}, test.Err)}
		success, err := mockClient.AliasChain("chain", "chain-alias")
		// if there is error as expected, the test passes
		if err != nil && test.Err != nil {
			continue
		}
		if err != nil {
			t.Fatalf("Unexepcted error: %s", err)
		}
		if success != test.Success {
			t.Fatalf("Expected success response to be: %v, but found: %v", test.Success, success)
		}
	}
}
